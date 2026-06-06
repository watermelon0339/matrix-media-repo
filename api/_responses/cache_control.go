package _responses

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/t2bot/matrix-media-repo/common/config"
)

const _IMMUTABLE_ETAG = "_IMMUTABLE_ETAG_MMR"

func SetMediaCacheControlHeaders(headers http.Header) {
	// 如果是本地开发，请使用下面这种设置（因为不走 CDN）
	// headers.Set("Cache-Control", "private, max-age=0, must-revalidate")

	cacheControlConfig := config.Get().CacheControl
	cacheControlHeader := fmt.Sprintf(
		"public, max-age=%d, s-maxage=%d, proxy-revalidate",
		cacheControlConfig.MaxAgeSeconds,
		cacheControlConfig.SMaxAgeSeconds,
	)
	// 如果是正式环境，请使用下面这种设置，配合 CDN 的缓存策略
	// `public` 允许 CDN 和浏览器缓存资源
	// `s-maxage=259200, proxy-revalidate` 告诉 CDN 缓存 3天，缓存过期之后回源验证资源是否有效
	// CDN(例如 Cloudflare) 的 Cache Rules 可以覆盖这个 `max-age` 和 `s-maxage`
	headers.Set("Cache-Control", cacheControlHeader)

	headers.Set("Etag", _IMMUTABLE_ETAG) // 资源是不可变的，使用固定的 Etag，允许 CDN 和浏览器通过 `If-None-Match` 来验证缓存
}

func CheckMediaCacheAndRespond(r *http.Request) interface{} {
	if matchesImmutableEtag(r.Header.Values("If-None-Match")) {
		return &NotModifiedResponse{}
	}

	return nil
}

func matchesImmutableEtag(values []string) bool {
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			candidate := strings.TrimSpace(part)
			if candidate == "" {
				continue
			}
			if candidate == "*" {
				return true
			}

			candidate = strings.TrimPrefix(candidate, "W/")
			candidate = strings.TrimSpace(candidate)
			candidate = strings.Trim(candidate, "\"")
			if candidate == _IMMUTABLE_ETAG {
				return true
			}
		}
	}

	return false
}
