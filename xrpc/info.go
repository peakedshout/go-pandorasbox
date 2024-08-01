package xrpc

import (
	"context"
	"encoding/json"
	"github.com/peakedshout/go-pandorasbox/tool/hjson"
	"net"
)

func CloneSessionAuthInfo(src, dst context.Context) context.Context {
	info := GetSessionAuthInfo(src)
	return SetSessionAuthInfo(dst, info.Clone())
}

func SetSessionAuthInfo(ctx context.Context, authInfo AuthInfo) context.Context {
	return context.WithValue(ctx, SessionAuthInfo, authInfo)
}

func SetSessionAuthInfoT[T any](ctx context.Context, k string, v T) (context.Context, error) {
	var authInfo AuthInfo
	if k == "" {
		err := hjson.UnmarshalV2(v, &authInfo)
		if err != nil {
			return nil, err
		}
		return context.WithValue(ctx, SessionAuthInfo, authInfo), nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var ok bool
	value := ctx.Value(SessionAuthInfo)
	if value == nil {
		authInfo = make(AuthInfo)
		ctx = context.WithValue(ctx, SessionAuthInfo, authInfo)
	} else if authInfo, ok = value.(AuthInfo); !ok {
		authInfo = make(AuthInfo)
		ctx = context.WithValue(ctx, SessionAuthInfo, authInfo)
	}
	authInfo.Set(k, string(b))
	return ctx, nil
}

func GetSessionAuthInfo(ctx context.Context) AuthInfo {
	value := ctx.Value(SessionAuthInfo)
	if value == nil {
		return make(AuthInfo)
	}
	if a, ok := value.(AuthInfo); ok {
		return a
	} else {
		return make(AuthInfo)
	}
}

func GetSessionAuthInfoT[T any](ctx context.Context, key string) (T, error) {
	info := GetSessionAuthInfo(ctx)
	var t T
	if key == "" {
		err := hjson.UnmarshalV2(info, &t)
		if err != nil {
			return t, err
		} else {
			return t, nil
		}
	} else {
		get := info.Get(key)
		err := json.Unmarshal([]byte(get), &t)
		if err != nil {
			return t, err
		} else {
			return t, nil
		}
	}
}

func CloneStreamAuthInfo(src, dst context.Context) context.Context {
	info := GetStreamAuthInfo(src)
	return SetStreamAuthInfo(dst, info.Clone())
}

func SetStreamAuthInfo(ctx context.Context, authInfo AuthInfo) context.Context {
	return context.WithValue(ctx, StreamAuthInfo, authInfo)
}

func SetStreamAuthInfoT[T any](ctx context.Context, k string, v T) (context.Context, error) {
	var authInfo AuthInfo
	if k == "" {
		err := hjson.UnmarshalV2(v, &authInfo)
		if err != nil {
			return nil, err
		}
		return context.WithValue(ctx, StreamAuthInfo, authInfo), nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var ok bool
	value := ctx.Value(StreamAuthInfo)
	if value == nil {
		authInfo = make(AuthInfo)
		ctx = context.WithValue(ctx, StreamAuthInfo, authInfo)
	} else if authInfo, ok = value.(AuthInfo); !ok {
		authInfo = make(AuthInfo)
		ctx = context.WithValue(ctx, StreamAuthInfo, authInfo)
	}
	authInfo.Set(k, string(b))
	return ctx, nil
}

func GetStreamAuthInfo(ctx context.Context) AuthInfo {
	value := ctx.Value(StreamAuthInfo)
	if value == nil {
		return make(AuthInfo)
	}
	if a, ok := value.(AuthInfo); ok {
		return a
	} else {
		return make(AuthInfo)
	}
}

func GetStreamAuthInfoT[T any](ctx context.Context, key string) (T, error) {
	info := GetStreamAuthInfo(ctx)
	var t T
	if key == "" {
		err := hjson.UnmarshalV2(info, &t)
		if err != nil {
			return t, err
		} else {
			return t, nil
		}
	} else {
		get := info.Get(key)
		err := json.Unmarshal([]byte(get), &t)
		if err != nil {
			return t, err
		} else {
			return t, nil
		}
	}
}

func GetAuthInfo[T any](info AuthInfo, key string) (T, error) {
	var t T
	if key == "" {
		err := hjson.UnmarshalV2(info, &t)
		if err != nil {
			return t, err
		} else {
			return t, nil
		}
	} else {
		get := info.Get(key)
		err := json.Unmarshal([]byte(get), &t)
		if err != nil {
			return t, err
		} else {
			return t, nil
		}
	}
}

func SetAuthInfo[T any](info AuthInfo, k string, v T) error {
	if k == "" {
		return hjson.UnmarshalV2(v, &info)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	info.Set(k, string(b))
	return nil
}

type AuthInfo map[string]string

func (a AuthInfo) Get(key string) string {
	return a[key]
}

func (a AuthInfo) Set(key, val string) {
	a[key] = val
}

func (a AuthInfo) Clone() AuthInfo {
	m := make(AuthInfo, len(a))
	for k, v := range a {
		m[k] = v
	}
	return a
}

func (a AuthInfo) connSet(s bool, conn net.Conn) AuthInfo {
	if s {
		a.Set(RemotePriAddress, hjson.MustMarshalStr(conn.LocalAddr().String()))
		a.Set(RemotePriNetwork, hjson.MustMarshalStr(conn.LocalAddr().Network()))
		a.Set(LocalPubAddress, hjson.MustMarshalStr(conn.RemoteAddr().String()))
		a.Set(LocalPubNetwork, hjson.MustMarshalStr(conn.RemoteAddr().Network()))
	} else {
		a.Set(RemotePubAddress, hjson.MustMarshalStr(conn.RemoteAddr().String()))
		a.Set(RemotePubNetwork, hjson.MustMarshalStr(conn.RemoteAddr().Network()))
		a.Set(LocalPriAddress, hjson.MustMarshalStr(conn.LocalAddr().String()))
		a.Set(LocalPriNetwork, hjson.MustMarshalStr(conn.LocalAddr().Network()))
	}
	return a
}

func bindsAuthInfo(ctx context.Context, ctxs ...context.Context) context.Context {
	info := GetSessionAuthInfo(ctx)
	for _, one := range ctxs {
		authInfo := GetSessionAuthInfo(one)
		for k, v := range authInfo {
			info[k] = v
		}
	}
	return context.WithValue(ctx, SessionAuthInfo, info)
}

type ConnInfo struct {
	RemotePubNetwork string
	RemotePubAddress string
	LocalPubNetwork  string
	LocalPubAddress  string

	RemotePriNetwork string
	RemotePriAddress string
	LocalPriNetwork  string
	LocalPriAddress  string
}

func GetConnInfo(ctx context.Context) ConnInfo {
	rpubnk, _ := GetSessionAuthInfoT[string](ctx, RemotePubNetwork)
	rpubad, _ := GetSessionAuthInfoT[string](ctx, RemotePubAddress)
	lpubnk, _ := GetSessionAuthInfoT[string](ctx, LocalPubNetwork)
	lpubad, _ := GetSessionAuthInfoT[string](ctx, LocalPubAddress)
	rprink, _ := GetSessionAuthInfoT[string](ctx, RemotePriNetwork)
	rpriad, _ := GetSessionAuthInfoT[string](ctx, RemotePriAddress)
	lprink, _ := GetSessionAuthInfoT[string](ctx, LocalPriNetwork)
	lpriad, _ := GetSessionAuthInfoT[string](ctx, LocalPriAddress)
	t := ConnInfo{
		RemotePubNetwork: rpubnk,
		RemotePubAddress: rpubad,
		LocalPubNetwork:  lpubnk,
		LocalPubAddress:  lpubad,
		RemotePriNetwork: rprink,
		RemotePriAddress: rpriad,
		LocalPriNetwork:  lprink,
		LocalPriAddress:  lpriad,
	}
	return t
}
