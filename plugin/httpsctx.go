package plugin

import "golang.org/x/net/context"

type key int

const httpsKey = -1000

func AddUseHttpsToContext(ctx context.Context, useCtx bool) context.Context {
	return context.WithValue(ctx, httpsKey, useCtx)
}

func GetUseHttpsContext(ctx context.Context) bool {
	useHttps, ok := ctx.Value(httpsKey).(bool)
	if !ok {
		return false
	}

	return useHttps
}
