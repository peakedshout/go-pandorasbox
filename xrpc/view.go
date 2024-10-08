package xrpc

import (
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"sort"
)

type SessionView struct {
	Id          string
	ConnInfo    ConnInfo
	MonitorInfo xnetutil.MonitorInfo
	StreamList  []StreamView
}

type StreamView struct {
	Id          string
	Type        string
	MonitorInfo xnetutil.MonitorInfo
}

func (s *Server) SessionView() []SessionView {
	var list []SessionView
	s.sessMap.Range(func(_ string, value *serverSession) bool {
		list = append(list, value.view())
		return true
	})
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].MonitorInfo.CreateTime.Before(list[j].MonitorInfo.CreateTime)
	})
	return list
}

func (ss *serverSession) view() SessionView {
	cinfo := GetConnInfo(ss.Context())
	sv := SessionView{
		Id:          ss.Id(),
		ConnInfo:    cinfo,
		MonitorInfo: ss.MonitorInfo(),
		StreamList:  nil,
	}
	ss.ssMux.Lock()
	sv.StreamList = make([]StreamView, 0, len(ss.streamMap)+len(ss.cacheMap))
	m := make(map[uint32]*serverStream, len(ss.streamMap)+len(ss.cacheMap))
	for id, stream := range ss.streamMap {
		m[id] = stream
	}
	for id, stream := range ss.cacheMap {
		m[id] = stream
	}
	for _, stream := range m {
		sv.StreamList = append(sv.StreamList, GetStreamView(stream))
	}
	ss.ssMux.Unlock()
	sort.SliceStable(sv.StreamList, func(i, j int) bool {
		return sv.StreamList[i].Id < sv.StreamList[j].Id
	})
	sort.SliceStable(sv.StreamList, func(i, j int) bool {
		return sv.StreamList[i].MonitorInfo.CreateTime.Before(sv.StreamList[j].MonitorInfo.CreateTime)
	})
	return sv
}

func (c *Client) SessionView() []SessionView {
	var list []SessionView
	c.sessMap.Range(func(_ string, value *ClientSession) bool {
		list = append(list, value.view())
		return true
	})
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].MonitorInfo.CreateTime.Before(list[j].MonitorInfo.CreateTime)
	})
	return list
}

func (cs *ClientSession) view() SessionView {
	cinfo := GetConnInfo(cs.Context())
	sv := SessionView{
		Id:          cs.Id(),
		ConnInfo:    cinfo,
		MonitorInfo: cs.xsess.MonitorInfo(),
		StreamList:  nil,
	}
	cs.streamMux.Lock()
	sv.StreamList = make([]StreamView, 0, len(cs.streamMap)+len(cs.cacheMap))
	m := make(map[uint32]*clientStream, len(cs.streamMap)+len(cs.cacheMap))
	for id, stream := range cs.streamMap {
		m[id] = stream
	}
	for id, stream := range cs.cacheMap {
		m[id] = stream
	}
	for _, stream := range m {
		sv.StreamList = append(sv.StreamList, GetStreamView(stream))
	}
	cs.streamMux.Unlock()
	sort.SliceStable(sv.StreamList, func(i, j int) bool {
		return sv.StreamList[i].Id < sv.StreamList[j].Id
	})
	sort.SliceStable(sv.StreamList, func(i, j int) bool {
		return sv.StreamList[i].MonitorInfo.CreateTime.Before(sv.StreamList[j].MonitorInfo.CreateTime)
	})
	return sv
}

func GetStreamView(stream Stream) StreamView {
	return StreamView{
		Id:          stream.Id(),
		Type:        stream.Id(),
		MonitorInfo: xnetutil.FormatMonitorInfo(stream),
	}
}
