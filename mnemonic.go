package main

import (
	"context"
	"fmt"
	"github.com/ernestosuarez/itertools"
	"github.com/miguelmota/go-ethereum-hdwallet"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	totalPermutations uint64 // 总排列数
	processedCount    uint64 // 已处理的排列数
	startTime         time.Time
	ctx               context.Context
	cancel            context.CancelFunc
)

// BruteForceResult 封装爆破结果
type BruteForceResult struct {
	Success     bool    // 是否成功
	Mnemonic    string  // 找到的助记词
	Address     string  // 对应的地址
	TimeElapsed string  // 耗时
	Processed   uint64  // 已处理的排列数
	Total       uint64  // 总排列数
	Rate        float64 // 处理速率（次/秒）
}

// calAddress 计算助记词对应的地址，并检查是否匹配目标地址
func calAddress(mnemonic, target string, resultChan chan<- string) {
	atomic.AddUint64(&processedCount, 1)
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return
	}
	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, false)
	if err != nil {
		return
	}
	address := account.Address.Hex()
	if address == target || strings.ToUpper(target) == strings.ToUpper(address) {
		resultChan <- fmt.Sprintf("%s", mnemonic) // 仅返回助记词
	}
}

// processSubset 处理以某个单词开头的所有排列组合
func processSubset(firstWord string, remainingWords []string, target string, wg *sync.WaitGroup, resultChan chan<- string, ctx context.Context) {
	defer wg.Done()

	// 计算排列数并更新总排列数
	permCount := factorial(len(remainingWords)) / factorial(len(remainingWords)-11)
	atomic.AddUint64(&totalPermutations, uint64(permCount))

	workerCount := runtime.NumCPU() * 8
	jobs := make(chan []string, workerCount)

	// 启动 worker
	var workerWg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for perm := range jobs {
				select {
				case <-ctx.Done():
					runtime.Goexit()
				default:
					mnemonic := firstWord + " " + strings.Join(perm, " ")
					calAddress(mnemonic, target, resultChan)
				}
			}
		}()
	}

	// 发送排列到工作池
	for perm := range itertools.PermutationsStr(remainingWords, 11) {
		select {
		case <-ctx.Done():
			close(jobs)
			workerWg.Wait()
			return
		case jobs <- perm:
		}
	}
	close(jobs)

	workerWg.Wait()
}

// factorial 计算阶乘（用于排列数计算）
func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * factorial(n-1)
}

func PrintProgress() string {
	processed := atomic.LoadUint64(&processedCount)
	total := atomic.LoadUint64(&totalPermutations)
	elapsed := time.Since(startTime)

	var progress float64
	if total > 0 {
		progress = float64(processed) / float64(total) * 100
	}

	rate := float64(processed) / elapsed.Seconds()
	return fmt.Sprintf("进度: %.4f%%, 已处理: %d/%d, 速率: %.2f 次/秒, 耗时: %v\n",
		progress, processed, total, rate, elapsed)
}

// BruteForceMnemonic 封装爆破逻辑，返回 BruteForceResult
func BruteForceMnemonic(words, target string) BruteForceResult {
	processedCount = 0
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	wordList := strings.Split(words, " ")
	var wg sync.WaitGroup
	resultChan := make(chan string, 1) // 缓冲为 1，避免阻塞
	startTime = time.Now()

	// 结果收集器
	var result BruteForceResult
	go func() {
		for mnemonic := range resultChan {
			result = BruteForceResult{
				Success:     true,
				Mnemonic:    mnemonic,
				Address:     target,
				TimeElapsed: time.Since(startTime).String(),
				Processed:   atomic.LoadUint64(&processedCount),
				Total:       atomic.LoadUint64(&totalPermutations),
				Rate:        float64(processedCount) / time.Since(startTime).Seconds(),
			}
			cancel()
			runtime.Goexit()
		}
	}()

	// 对每个单词作为开头进行并行处理
	for i, firstWord := range wordList {
		wg.Add(1)
		remainingWords := make([]string, 0, len(wordList)-1)
		remainingWords = append(remainingWords, wordList[:i]...)
		remainingWords = append(remainingWords, wordList[i+1:]...)

		go func(fw string, rw []string) {
			processSubset(fw, rw, target, &wg, resultChan, ctx)
		}(firstWord, remainingWords)
	}

	wg.Wait()
	close(resultChan)
	// 如果没有找到结果，返回默认值
	if !result.Success {
		result = BruteForceResult{
			Success:     false,
			TimeElapsed: time.Since(startTime).String(),
			Processed:   atomic.LoadUint64(&processedCount),
			Total:       atomic.LoadUint64(&totalPermutations),
			Rate:        float64(processedCount) / time.Since(startTime).Seconds(),
		}
	}

	return result
}

func StopBruteForceMnemonic() {
	cancel()
}
