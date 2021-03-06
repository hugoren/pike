package cache

import (
	"bytes"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/vicanso/pike/vars"

	"github.com/valyala/fasthttp"
)

func TestInitRequestStatus(t *testing.T) {
	rs := initRequestStatus(30)
	if rs.createdAt == 0 {
		t.Fatalf("the created at should be seconds for now")
	}
	if rs.ttl != 30 {
		t.Fatalf("the ttl should be 30s")
	}
}

func TestIsExpired(t *testing.T) {
	rs := initRequestStatus(30)
	if isExpired(rs) != false {
		t.Fatalf("the rs should not be expired")
	}
	rs.createdAt = 0
	if isExpired(rs) != true {
		t.Fatalf("the rs should be expired")
	}
}

func TestByteToUnit(t *testing.T) {
	b16 := uint16ToBytes(100)
	v16 := bytesToUint16(b16)
	if v16 != 100 {
		t.Fatalf("the uint16 to bytes fail")
	}

	b32 := uint32ToBytes(100)
	v32 := bytesToUint32(b32)
	if v32 != 100 {
		t.Fatalf("the uint32 to bytes fail")
	}
}

func TestTrimHeader(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	header := &ctx.Request.Header
	header.SetCanonical([]byte("Content-Type"), []byte("application/json; charset=utf-8"))
	header.SetCanonical([]byte("X-Response-Id"), []byte("BJJRAyf4f"))
	header.SetCanonical([]byte("Cache-Control"), []byte("no-cache, max-age=0"))
	header.SetCanonical([]byte("Connection"), []byte("keep-alive"))
	header.SetCanonical([]byte("Date"), []byte("Tue, 09 Jan 2018 12:27:02 GMT"))
	str := "User-Agent: fasthttp\r\nContent-Type: application/json; charset=utf-8\r\nX-Response-Id: BJJRAyf4f\r\nCache-Control: no-cache, max-age=0"
	data := string(trimHeader(header.Header()))
	if data != str {
		t.Fatalf("trim header fail expect %v but %v", str, data)
	}
}

func TestDB(t *testing.T) {
	_, err := InitDB("/tmp/pike")
	if err != nil {
		t.Fatalf("open db fail, %v", err)
	}
	key := []byte("/users/me")
	data := []byte("vicanso")
	err = Save(key, data, 200)
	if err != nil {
		t.Fatalf("save data fail %v", err)
	}
	buf, err := Get(key)
	if err != nil {
		t.Fatalf("get data fail %v", err)
	}
	if bytes.Compare(data, buf) != 0 {
		t.Fatalf("get data fail")
	}

	ctx := &fasthttp.RequestCtx{}
	ctx.Response.SetBody(data)
	ctx.Response.Header.SetCanonical(vars.CacheControl, []byte("public, max-age=30"))

	resBody := ctx.Response.Body()
	resHeader := ctx.Response.Header.Header()

	saveRespData := &ResponseData{
		CreatedAt:      uint32(time.Now().Unix()),
		StatusCode:     200,
		Compress:       vars.GzipData,
		ShouldCompress: true,
		TTL:            30,
		Header:         resHeader,
		Body:           resBody,
	}

	SaveResponseData(key, saveRespData)
	respData, err := GetResponse(key)
	if err != nil {
		t.Fatalf("get the response fail, %v", err)
	}
	if uint32(time.Now().Unix())-respData.CreatedAt > 1 {
		t.Fatalf("get the create time fail")
	}
	if respData.TTL != 30 {
		t.Fatalf("get the ttle fail")
	}
	checkHeader := []byte("Server: fasthttp\r\nCache-Control: public, max-age=30")
	if bytes.Compare(respData.Header, checkHeader) != 0 {
		t.Fatalf("the response header fail")
	}
	if bytes.Compare(respData.Body, data) != 0 {
		t.Fatalf("the response body fail")
	}
	if respData.Compress != vars.GzipData {
		t.Fatalf("the data should be gzip compress")
	}
	if !respData.ShouldCompress {
		t.Fatalf("the data should be compress")
	}
}

func TestRequestStatus(t *testing.T) {
	key := []byte("GEThttp://aslant.site/users/me")
	status, c := GetRequestStatus(key)
	// 第一次请求时，状态为fetching
	if status != vars.Fetching {
		t.Fatalf("the first request should be fetching")
	}
	if c != nil {
		t.Fatalf("the chan of first request should be nil")
	}

	status, c = GetRequestStatus(key)
	if status != vars.Waiting {
		t.Fatalf("the second request should be wating for the first request result")
	}
	if c == nil {
		t.Fatalf("the chan of second request shouldn't be nil")
	}
	go func(tmp chan int) {
		tmpStatus := <-tmp
		if tmpStatus != vars.HitForPass {
			t.Fatalf("the waiting request should be hit for pass")
		}
	}(c)

	HitForPass(key, 100)
	time.Sleep(time.Second)

	key = []byte("GEThttp://aslant.site/books")
	status, c = GetRequestStatus(key)
	// 第一次请求时，状态为fetching
	if status != vars.Fetching {
		t.Fatalf("the first request should be fetching")
	}
	if c != nil {
		t.Fatalf("the chan of first request should be nil")
	}

	status, c = GetRequestStatus(key)
	if status != vars.Waiting {
		t.Fatalf("the second request should be wating for the first request result")
	}
	if c == nil {
		t.Fatalf("the chan of second request shouldn't be nil")
	}
	go func(tmp chan int) {
		tmpStatus := <-tmp
		if tmpStatus != vars.Cacheable {
			t.Fatalf("the waiting request should be cacheable")
		}
	}(c)

	Cacheable(key, 100)
	size := Size()
	if size != 2 {
		t.Fatalf("the cache size expect 2 but %v", size)
	}
	lsm, vLog := DataSize()
	if lsm == -1 || vLog == -1 {
		t.Fatalf("get the data size fail")
	}
	fetchingCount, waitingCount, cacheableCount, hitForPassCount := Stats()
	if fetchingCount != 0 || waitingCount != 0 {
		t.Fatalf("the fecthing and wating count is wrong")
	}
	if cacheableCount != 1 {
		t.Fatalf("the cacheable count expect 1 but %v", cacheableCount)
	}
	if hitForPassCount != 1 {
		t.Fatalf("the hit for pass count expect 1 but %v", hitForPassCount)
	}
	time.Sleep(time.Second)
	buf := GetCachedList()
	if bytes.Index(buf, []byte("GEThttp://aslant.site/books")) == -1 {
		t.Fatalf("the cache list should include GEThttp://aslant.site/books")
	}

	status, _ = GetRequestStatus(key)
	if status != vars.Cacheable {
		t.Fatalf("the %s should be cacheable", key)
	}
	Expire(key)
	status, _ = GetRequestStatus(key)
	if status != vars.Fetching {
		t.Fatalf("the %s should be fetching", key)
	}
}

func TestResponseCache(t *testing.T) {
	// 测试生成插入多条记录，将对过期数据删除
	startedAt := time.Now()
	count := 10 * 1024
	for index := 0; index < count; index++ {
		key := []byte("test-" + strconv.Itoa(index))
		SaveResponseData(key, &ResponseData{
			CreatedAt:  uint32(time.Now().Unix()),
			StatusCode: 200,
			Compress:   vars.RawData,
			TTL:        10,
			Header:     make([]byte, 3*1024),
			Body:       make([]byte, 50*1024),
		})
	}
	log.Printf("create %v use %v", count, time.Since(startedAt))
	ClearExpired()
	Close()
}
