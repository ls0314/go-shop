// loadgen 轻量压测工具(DS-A-19 压测专用)
//
// 用途:对领券等热点接口做并发施压,输出 QPS / 延迟分位数 / 状态码分布。
// 为什么不用 hey:Windows 下 MSYS 对 -H/-d 参数的转义存在不确定性(出现过
// body 丢失导致 400、头部异常导致 401 的混合现象);本工具原生编译,行为确定,
// 且延迟统计直接内建,不依赖外部解析。
//
// 用法:
//
//	go run ./cmd/loadgen -url http://... -token <JWT> -z 30s -c 200
//	go run ./cmd/loadgen -url http://... -token <JWT> -n 2000 -c 200
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := flag.String("url", "", "目标 URL(必填)")
	token := flag.String("token", "", "Bearer token")
	body := flag.String("body", "{}", "请求体")
	dur := flag.Duration("z", 0, "时长模式,如 30s;与 -n 二选一")
	total := flag.Int64("n", 0, "总请求数模式;与 -z 二选一")
	conc := flag.Int("c", 50, "并发连接数")
	flag.Parse()
	if *url == "" || (*dur == 0 && *total == 0) {
		flag.Usage()
		os.Exit(1)
	}

	client := &http.Client{
		Transport: &http.Transport{
			// 连接池容量对齐并发数:复用 keep-alive 连接,压测的是服务端而非建连
			MaxIdleConns:        *conc,
			MaxIdleConnsPerHost: *conc,
			MaxConnsPerHost:     0,
		},
	}

	var sent, done, failed atomic.Int64
	statusCount := sync.Map{} // status code -> *atomic.Uint64
	latencies := make([][]float64, *conc)
	deadline := time.Time{}
	if *dur > 0 {
		deadline = time.Now().Add(*dur)
	}

	var wg sync.WaitGroup
	start := time.Now()
	for w := 0; w < *conc; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for {
				if !deadline.IsZero() && time.Now().After(deadline) {
					return
				}
				if *total > 0 {
					if sent.Add(1) > *total {
						return
					}
				}
				req, err := http.NewRequest(http.MethodPost, *url, strings.NewReader(*body))
				if err != nil {
					failed.Add(1)
					continue
				}
				req.Header.Set("Authorization", "Bearer "+*token)
				req.Header.Set("Content-Type", "application/json")

				t0 := time.Now()
				resp, err := client.Do(req)
				elapsed := time.Since(t0).Seconds() * 1000
				if err != nil {
					failed.Add(1)
					continue
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				latencies[w] = append(latencies[w], elapsed)
				done.Add(1)
				key := resp.StatusCode
				v, _ := statusCount.LoadOrStore(key, &atomic.Uint64{})
				v.(*atomic.Uint64).Add(1)
			}
		}(w)
	}
	wg.Wait()
	elapsed := time.Since(start)

	// 汇总延迟分位数
	var all []float64
	for _, l := range latencies {
		all = append(all, l...)
	}
	sort.Float64s(all)
	pct := func(p float64) string {
		if len(all) == 0 {
			return "n/a"
		}
		idx := int(float64(len(all)-1) * p)
		return fmt.Sprintf("%.1fms", all[idx])
	}

	fmt.Println("========== loadgen 结果 ==========")
	fmt.Printf("目标: %s\n并发: %d  完成: %d  网络失败: %d  耗时: %.1fs\n",
		*url, *conc, done.Load(), failed.Load(), elapsed.Seconds())
	fmt.Printf("QPS: %.0f\n", float64(done.Load())/elapsed.Seconds())
	fmt.Printf("p50=%s  p90=%s  p99=%s  max=%s\n",
		pct(0.50), pct(0.90), pct(0.99), pct(1.0))
	fmt.Println("-- 状态码分布 --")
	statusCount.Range(func(k, v interface{}) bool {
		fmt.Printf("  [%s] %d\n", strconv.Itoa(k.(int)), v.(*atomic.Uint64).Load())
		return true
	})
}
