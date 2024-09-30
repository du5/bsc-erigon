package jsonrpc

import (
	"context"
	"net"
	"strings"
	"time"

	redis_sdk "github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type server struct {
	*redis_sdk.Client
}

var redis server

func (s server) addIPToBlockList(ctxValue any) error {
	ip := s.getIPFromContext(ctxValue)
	if ip == "" {
		return nil
	}

	cmd := s.Client.Set(ctx, ip, "", time.Minute*5)
	return cmd.Err()
}

func (s server) getIPFromContext(v any) string {
	if v == nil {
		return ""
	}

	ip, ok := v.(string)
	if !ok {
		return ""
	}
	return formatIP(ip)
}

func (s server) isIPBlocked(ctxValue any) bool {
	ip := s.getIPFromContext(ctxValue)
	if ip == "" {
		return false
	}

	cmd := s.Client.Exists(ctx, ip)
	if cmd.Err() != nil {
		return false
	}
	return cmd.Val() == 1
}

func formatIP(s string) string { // check ipv4, or format ipv6 to cidr64
	if strings.Index(s, ":") > 0 {
		_, cidr64, err := net.ParseCIDR(s + "/64") // format ipv6 to cidr64, if bad ip, return empty string and toBlockIP
		if err != nil {
			return ""
		}
		return cidr64.String()
	}

	ip := net.ParseIP(s) // bad ip will be nil, return empty string and toBlockIP
	if ip == nil || ip.String() != s {
		return ""
	}

	return s
}

func init() {
	redis = server{redis_sdk.NewClient(&redis_sdk.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})}
}
