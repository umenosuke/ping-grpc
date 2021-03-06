package main

import (
	"context"
	"sync"
	"time"

	"github.com/umenosuke/labelinglog"
	"github.com/umenosuke/pinger4"

	pb "github.com/umenosuke/ping-grpc-server/proto/pingGrpc"
)

type tPingerWrap struct {
	pinger            *pinger4.Pinger
	idStr             string
	startUnixNanosec  uint64
	expireUnixNanosec uint64
	description       string
	cancelFunc        context.CancelFunc
	chResultListener  struct {
		sync.Mutex
		list []chan<- *pb.IcmpResult
	}
	chStatisticsListener struct {
		sync.Mutex
		list []chan<- *pb.Statistics
	}
	statisticsInterval uint64
}

func (thisPingerWrap *tPingerWrap) start(ctx context.Context) {
	wgChild := sync.WaitGroup{}

	wgChild.Add(1)
	go (func() {
		defer wgChild.Done()
		defer thisPingerWrap.cancelFunc()
		defer logger.Log(labelinglog.FlgDebug, "(id "+thisPingerWrap.idStr+")"+" finish PingerWrap Run")
		logger.Log(labelinglog.FlgDebug, "(id "+thisPingerWrap.idStr+")"+" Start PingerWrap Run")

		thisPingerWrap.pinger.Run(ctx)
	})()

	wgChild.Add(1)
	go (func() {
		defer wgChild.Done()
		defer logger.Log(labelinglog.FlgDebug, "(id "+thisPingerWrap.idStr+")"+" finish PingerWrap result")
		logger.Log(labelinglog.FlgDebug, "(id "+thisPingerWrap.idStr+")"+" Start PingerWrap result")
		thisPingerWrap.result(ctx)
	})()

	wgChild.Add(1)
	go (func() {
		defer wgChild.Done()
		defer logger.Log(labelinglog.FlgDebug, "(id "+thisPingerWrap.idStr+")"+" finish PingerWrap statistics")
		logger.Log(labelinglog.FlgDebug, "(id "+thisPingerWrap.idStr+")"+" Start PingerWrap statistics")
		thisPingerWrap.statistics(ctx)
	})()

	wgChild.Wait()
}

func (thisPingerWrap *tPingerWrap) addResultListener(ch chan<- *pb.IcmpResult) {
	thisPingerWrap.chResultListener.Lock()
	defer thisPingerWrap.chResultListener.Unlock()
	if thisPingerWrap.chResultListener.list != nil {
		thisPingerWrap.chResultListener.list = append(thisPingerWrap.chResultListener.list, ch)
	} else {
		close(ch)
	}
}

func (thisPingerWrap *tPingerWrap) result(ctx context.Context) {
	defer thisPingerWrap.cancelFunc()
	defer logger.Log(labelinglog.FlgDebug, "(id "+thisPingerWrap.idStr+")"+" finish pinger.GetChIcmpResult")
	defer (func() {
		thisPingerWrap.chResultListener.Lock()
		defer thisPingerWrap.chResultListener.Unlock()
		for _, ch := range thisPingerWrap.chResultListener.list {
			close(ch)
		}
		thisPingerWrap.chResultListener.list = nil
	})()
	logger.Log(labelinglog.FlgDebug, "(id "+thisPingerWrap.idStr+")"+" Start pinger.GetChIcmpResult")

	targetsOrder := thisPingerWrap.pinger.GetInfo().TargetsOrder
	chIcmpResult := thisPingerWrap.pinger.GetChIcmpResult(len(targetsOrder) * 2)
	for {
		select {
		case <-ctx.Done():
			return
		case result := <-chIcmpResult:
			var resType pb.IcmpResult_ResultType
			switch result.ResultType {
			case pinger4.IcmpResultTypeReceive:
				resType = pb.IcmpResult_IcmpResultTypeReceive
			case pinger4.IcmpResultTypeReceiveAfterTimeout:
				resType = pb.IcmpResult_IcmpResultTypeReceiveAfterTimeout
			case pinger4.IcmpResultTypeTTLExceeded:
				resType = pb.IcmpResult_IcmpResultTypeTTLExceeded
			case pinger4.IcmpResultTypeTimeout:
				resType = pb.IcmpResult_IcmpResultTypeTimeout
			default:
				resType = pb.IcmpResult_IcmpResultTypeUnknown
			}
			pbResult := pb.IcmpResult{
				Type:                   resType,
				TargetID:               uint32(result.IcmpTargetID),
				BinPeerIP:              uint32(result.BinPeerIP),
				Sequence:               int64(result.Seq),
				SendTimeUnixNanosec:    int64(result.SendTimeUnixNanosec),
				ReceiveTimeUnixNanosec: int64(result.ReceiveTimeUnixNanosec),
			}

			(func() {
				thisPingerWrap.chResultListener.Lock()
				defer thisPingerWrap.chResultListener.Unlock()
				for _, ch := range thisPingerWrap.chResultListener.list {
					select {
					case ch <- &pbResult:
					default:
					}
				}
			})()
		}
	}
}

func (thisPingerWrap *tPingerWrap) addStatisticsListener(ch chan<- *pb.Statistics) {
	thisPingerWrap.chStatisticsListener.Lock()
	defer thisPingerWrap.chStatisticsListener.Unlock()
	if thisPingerWrap.chStatisticsListener.list != nil {
		thisPingerWrap.chStatisticsListener.list = append(thisPingerWrap.chStatisticsListener.list, ch)
	} else {
		close(ch)
	}
}

func (thisPingerWrap *tPingerWrap) statistics(ctx context.Context) {
	defer thisPingerWrap.cancelFunc()
	defer logger.Log(labelinglog.FlgDebug, "(id "+thisPingerWrap.idStr+")"+" finish pinger.GetStatistics")
	defer (func() {
		thisPingerWrap.chStatisticsListener.Lock()
		defer thisPingerWrap.chStatisticsListener.Unlock()
		for _, ch := range thisPingerWrap.chStatisticsListener.list {
			close(ch)
		}
		thisPingerWrap.chStatisticsListener.list = nil
	})()
	logger.Log(labelinglog.FlgDebug, "(id "+thisPingerWrap.idStr+")"+" Start pinger.GetStatistics")

	interval := time.Duration(thisPingerWrap.statisticsInterval) * time.Second

	targetsOrder := thisPingerWrap.pinger.GetInfo().TargetsOrder
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
			counts := thisPingerWrap.pinger.GetSuccessCounts()
			pbCounts := make([]*pb.Statistics_SuccessCount, 0)
			for _, id := range targetsOrder {
				pbCounts = append(pbCounts, &pb.Statistics_SuccessCount{
					TargetID: uint32(id),
					Count:    counts[id].Count,
				})
			}

			pbStatistics := &pb.Statistics{
				Targets: pbCounts,
			}

			(func() {
				thisPingerWrap.chStatisticsListener.Lock()
				defer thisPingerWrap.chStatisticsListener.Unlock()
				for _, ch := range thisPingerWrap.chStatisticsListener.list {
					select {
					case ch <- pbStatistics:
					default:
					}
				}
			})()
		}
	}
}
