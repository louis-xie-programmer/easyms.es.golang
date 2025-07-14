package easyes

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"easyms-es/config"
	"easyms-es/fasthttp"
	"easyms-es/model"
	"easyms-es/utility"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// StoreConfig Store存储配置参数
type StoreConfig struct {
	IndexName string
	Timeout   time.Duration
}

// Store Store对象
type Store struct {
	es        *elasticsearch.Client
	IndexName string
}

// NewStore 构造
func NewStore(c StoreConfig) (*Store, error) {
	var (
		caCertPath = config.GetSyncConfig("", "common.elasticsearch.cacert")
		esAddress  = config.GetSyncConfig("", "common.elasticsearch.address")
		esUserName = config.GetSyncConfig("", "common.elasticsearch.username")
		esPassword = config.GetSyncConfig("", "common.elasticsearch.password")
	)

	indexName := c.IndexName
	if indexName == "" {
		indexName = model.EsProductIndexName
	}

	// read es cert file
	cert, err := ioutil.ReadFile(caCertPath)

	if err != nil {
		//log.Fatalln("common.elasticsearch.cacert error:", err)
		return nil, fmt.Errorf("es error read elasticsearch cert failed: %v", err)
	}

	// elasticsearch client 默认net/http
	// 修改Transport 中的MaxIdleConnsPerHost, 以便处理http不被重用的问题(time_wait)
	/*
		cfg := elasticsearch.Config{
			Addresses: strings.Split(esAddress, ","),
			Username:  esUserName,
			Password:  esPassword,
			CACert:    cert,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   c.Timeout,
					KeepAlive: c.Timeout,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				MaxConnsPerHost:       100,
				MaxIdleConnsPerHost:   100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}
	*/

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cert)

	cfg := elasticsearch.Config{
		Addresses: strings.Split(esAddress, ","),
		Username:  esUserName,
		Password:  esPassword,
		Transport: fasthttp.NewTransport(&tls.Config{
			RootCAs:    caCertPool,
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
		}),
	}

	es, err := elasticsearch.NewClient(cfg)

	if err != nil {
		return nil, fmt.Errorf("es error creating the client: %s", err.Error())
	}

	s := Store{es: es, IndexName: indexName}
	return &s, nil
}

// CreateIndex 通过mapping创建索引
func (s *Store) CreateIndex(mapping string) error {
	res, err := s.es.Indices.Create(s.IndexName, s.es.Indices.Create.WithBody(strings.NewReader(mapping)))
	if err != nil {
		return fmt.Errorf("es error creating index: %s", err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}(res.Body)
	if res.IsError() {
		return fmt.Errorf("es error: [%s] [%s] %s", res.Status(), s.IndexName, mapping)
	}
	return nil
}

// Create 插入一条索引数据
func (s *Store) Create(item interface{}) error {
	payload, err := json.Marshal(item)
	if err != nil {
		return err
	}

	_id := utility.GetFieldIDTag(item)
	if _id == "" {
		return fmt.Errorf("bulk obj id is error")
	}
	documentId := _id

	ctx := context.Background()
	res, err := esapi.CreateRequest{
		Index:      s.IndexName,
		DocumentID: documentId,
		Body:       bytes.NewReader(payload),
	}.Do(ctx, s.es)
	if err != nil {
		return fmt.Errorf("es error insert doc request: %s", err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}(res.Body)

	if res.IsError() {
		return fmt.Errorf("es error [%s] [%s] %s", res.Status(), s.IndexName, res.Body)
	}

	return nil
}

// Bulk 批量插入,更新,删除   对象要进行处理_id赋值,注意struct 中对tag id
func (s *Store) Bulk(items interface{}, flag ...string) error {
	opt := "index"
	if len(flag) > 0 {
		opt = flag[0]
	}

	array, ok := items.([]any)
	if !ok || len(array) < 1 {
		return fmt.Errorf("data is not array BulkInsert error.")
	}

	var buf bytes.Buffer

	for _, item := range array {
		_id := utility.GetFieldIDTag(item)
		if _id == "" {
			return fmt.Errorf("bulk obj id is error")
		}
		documentId := _id

		meta := []byte(fmt.Sprintf(`{ "%s" : { "_id" : "%s" } }%s`, opt, documentId, "\n"))
		if opt != "delete" {
			var data []byte
			if opt == "update" {
				data, _ = json.Marshal(map[string]interface{}{
					"doc": item,
				})
			} else {
				data, _ = json.Marshal(item)
			}

			data = append(data, "\n"...)
			buf.Grow(len(meta) + len(data))
			buf.Write(meta)
			buf.Write(data)
		} else {
			buf.Grow(len(meta))
			buf.Write(meta)
		}
	}

	res, err := s.es.Bulk(bytes.NewReader(buf.Bytes()), s.es.Bulk.WithIndex(s.IndexName))

	if err != nil {
		return fmt.Errorf("es error bulk insert docs request: %s", err.Error())
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}(res.Body)

	type bulkResponse struct {
		Errors bool `json:"errors"`
		Items  []struct {
			Index struct {
				ID     string `json:"_id"`
				Result string `json:"result"`
				Status int    `json:"status"`
				Error  struct {
					Type   string `json:"type"`
					Reason string `json:"reason"`
					Cause  struct {
						Type   string `json:"type"`
						Reason string `json:"reason"`
					} `json:"caused_by"`
				} `json:"error"`
			} `json:"index"`
		} `json:"items"`
	}

	var blk *bulkResponse

	if res.IsError() {
		var raw map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
			//log.Fatalf("Failure to to parse response body: %s", err)
			return fmt.Errorf("Failure to to parse response body: %s", err.Error())
		}
		return fmt.Errorf("[%s] %s: %s", res.Status(), raw["error"].(map[string]interface{})["type"], raw["error"].(map[string]interface{})["reason"])
	} else {
		if err := json.NewDecoder(res.Body).Decode(&blk); err != nil {
			//log.Fatalf("Failure to to parse response body: %s", err)
			return fmt.Errorf("Failure to to parse response body: %s", err)
		} else {
			for _, d := range blk.Items {
				// ... so for any HTTP status above 201 ...
				//
				if d.Index.Status > 201 {
					// ... and print the response status and error information ...
					log.Printf("  Error: [%d]: %s: %s: %s: %s",
						d.Index.Status,
						d.Index.Error.Type,
						d.Index.Error.Reason,
						d.Index.Error.Cause.Type,
						d.Index.Error.Cause.Reason,
					)
				}
			}
		}
	}

	defer buf.Reset()

	return nil
}

// Get 通过DocumentID获取一条数据
func (s *Store) Get(id string) (*DocResponse, error) {
	req := esapi.GetRequest{
		Index:      s.IndexName,
		DocumentID: id,
	}
	res, err := req.Do(context.Background(), s.es)
	if err != nil {
		return nil, fmt.Errorf("es error getting doc: %s", err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}(res.Body)
	if res.IsError() {
		return nil, fmt.Errorf("es error [%s] [%s] %s ID NOT EXIST.", res.Status(), s.IndexName, id)
	}
	var r DocResponse
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	return &r, nil
}

// Exists 查询DocumentID是否存在
func (s *Store) Exists(id string) (bool, error) {
	res, err := s.es.Exists(s.IndexName, id)
	if err != nil {
		return false, fmt.Errorf("es error Exists doc: %s", err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}(res.Body)
	switch res.StatusCode {
	case 200:
		return true, nil
	case 404:
		return false, nil
	default:
		return false, fmt.Errorf("es error Exists error [%s] [%s] %s", res.Status(), s.IndexName, id)
	}
}

// Count 获取查询条件的数量
func (s *Store) Count(body string) (int, error) {
	res, err := s.es.Count(
		s.es.Count.WithContext(context.Background()),
		s.es.Count.WithIndex(s.IndexName),
		s.es.Count.WithBody(strings.NewReader(body)),
	)
	if err != nil {
		return 0, fmt.Errorf("es error Count: %s", err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}(res.Body)

	if res.IsError() {
		return 0, fmt.Errorf("es error [%s] [%s] %s", res.Status(), s.IndexName, body)
	}

	type CountResponse struct {
		Count int `json:"count"`
	}

	var r CountResponse
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return 0, err
	}

	return r.Count, nil
}

// Search 搜索
func (s *Store) Search(body string) (*SearchResponse, error) {
	res, err := s.es.Search(
		s.es.Search.WithContext(context.Background()),
		s.es.Search.WithIndex(s.IndexName),
		s.es.Search.WithBody(strings.NewReader(body)),
		s.es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("es error Search: %s", err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}(res.Body)

	if res.IsError() {
		return nil, fmt.Errorf("es error [%s] [%s] %s", res.Status(), s.IndexName, body)
	}

	var r SearchResponse
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	return &r, nil
}

// SearchGradeAgg 多级分组聚合
func (s *Store) SearchGradeAgg(body string) (*GradeAggResult, error) {
	res, err := s.es.Search(
		s.es.Search.WithContext(context.Background()),
		s.es.Search.WithIndex(s.IndexName),
		s.es.Search.WithBody(strings.NewReader(body)),
	)
	if err != nil {
		return nil, fmt.Errorf("es error SearchGradeAgg: %s", err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}(res.Body)

	if res.IsError() {
		return nil, fmt.Errorf("es error [%s] [%s] %s", res.Status(), s.IndexName, body)
	}

	var r GradeAggResult
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	return &r, nil
}

// SearchAgg 搜索普通聚合
func (s *Store) SearchAgg(body string) (*AggResult, error) {
	res, err := s.es.Search(
		s.es.Search.WithContext(context.Background()),
		s.es.Search.WithIndex(s.IndexName),
		s.es.Search.WithBody(strings.NewReader(body)),
	)
	if err != nil {
		return nil, fmt.Errorf("es error SearchAgg: %s", err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}(res.Body)

	if res.IsError() {
		return nil, fmt.Errorf("es error [%s] [%s] %s", res.Status(), s.IndexName, body)
	}

	var r AggResult
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	return &r, nil
}

// SearchAggMultiple 并发聚合（多个聚合同时搜索）
func (s *Store) SearchAggMultiple(body string) (*[]AggResult, error) {
	req := esapi.MsearchRequest{
		Body:   strings.NewReader(body),
		Pretty: true,
	}

	res, err := req.Do(context.Background(), s.es)
	if err != nil {
		return nil, fmt.Errorf("es error SearchAggMultiple: %s", err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}(res.Body)

	if res.IsError() {
		return nil, fmt.Errorf("es error [%s] [%s] %s", res.Status(), s.IndexName, body)
	}

	type MResponses struct {
		Responses []AggResult
	}

	var r MResponses
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	return &r.Responses, nil
}

// SearchSamplerAggMultiple 采样聚合(多)
func (s *Store) SearchSamplerAggMultiple(body string, aggNames []string) (*[]AggResult, error) {
	req := esapi.MsearchRequest{
		Body:   strings.NewReader(body),
		Pretty: true,
	}

	res, err := req.Do(context.Background(), s.es)
	if err != nil {
		return nil, fmt.Errorf("es error SearchSamplerAggMultiple: %s", err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}(res.Body)

	if res.IsError() {
		return nil, fmt.Errorf("es error [%s] [%s] %s", res.Status(), s.IndexName, body)
	}

	type MResponses struct {
		Responses []struct {
			Aggregations struct {
				Sample struct {
					DocCount int `json:"doc_count,omitempty"`
					fields   map[string]Aggregation
				}
			}
		}
	}

	var r MResponses
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	var rst []AggResult
	for _, rsp := range r.Responses {
		var agg AggResult
		agg.Aggregations = make(map[string]Aggregation)
		if rsp.Aggregations.Sample.DocCount < 1 {
			continue
		}
		for _, aggName := range aggNames {
			agg.Aggregations[aggName] = rsp.Aggregations.Sample.fields[aggName]
		}
		rst = append(rst, agg)
	}

	return &rst, nil
}

// SearchCollapse 折叠去重
func (s *Store) SearchCollapse(body string) (*CollapseSearchResponse, error) {
	res, err := s.es.Search(
		s.es.Search.WithContext(context.Background()),
		s.es.Search.WithIndex(s.IndexName),
		s.es.Search.WithBody(strings.NewReader(body)),
	)
	if err != nil {
		return nil, fmt.Errorf("es error SearchCollapse: %s", err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}(res.Body)

	if res.IsError() {
		return nil, fmt.Errorf("es error [%s] [%s] %s", res.Status(), s.IndexName, body)
	}

	var r CollapseSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	return &r, nil
}

// Analyze 分析
func (s *Store) Analyze(analyzer string, text string) (*Tokens, error) {
	dsl := fmt.Sprintf(`{
		"analyzer" : "%s",
		"text": "%s"
		}`, analyzer, text)

	ares, err := s.es.Indices.Analyze(
		s.es.Indices.Analyze.WithContext(context.Background()),
		s.es.Indices.Analyze.WithIndex(s.IndexName),
		s.es.Indices.Analyze.WithBody(strings.NewReader(dsl)),
	)
	if err != nil {
		return nil, fmt.Errorf("es error Analyze request: %s", err.Error())
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}(ares.Body)

	if ares.IsError() {
		return nil, fmt.Errorf("es error [%s] [%s] %s", ares.Status(), s.IndexName, dsl)
	}

	var r Tokens
	if err := json.NewDecoder(ares.Body).Decode(&r); err != nil {
		return nil, err
	}

	return &r, nil
}

// UpdateByQuery query to update document  批量按照条件更新document
func (s *Store) UpdateByQuery(body string) (string, error) {
	res, err := s.es.UpdateByQuery(
		[]string{s.IndexName},
		s.es.UpdateByQuery.WithContext(context.Background()),
		s.es.UpdateByQuery.WithBody(strings.NewReader(body)),
		s.es.UpdateByQuery.WithConflicts("proceed"),
		s.es.UpdateByQuery.WithWaitForCompletion(false),
		s.es.UpdateByQuery.WithPretty(),
	)
	if err != nil {
		return "", fmt.Errorf("es error UpdateByQuery: %s", err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("failed to close stmt: %v", err)
		}
	}(res.Body)
	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return "", err
	}
	taskID := response["task"].(string)

	return taskID, nil
}

// MonitorTask 监控task完成状态
func (s *Store) MonitorTask(taskID string) (bool, error) {
	for {
		err := func() error {
			// 获取任务状态
			res, err := s.es.Tasks.Get(taskID)
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					log.Printf("failed to close stmt: %v", err)
				}
			}(res.Body)

			// 解析任务状态
			var taskStatus map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&taskStatus); err != nil {
				return err
			}

			// 检查任务是否完成
			if !taskStatus["completed"].(bool) {
				time.Sleep(5 * time.Second) // 等待一段时间后再次检查
			}
			return nil
		}()
		if err != nil {
			return false, fmt.Errorf("es error MonitorTask: %s", err.Error())
		}

	}
}

// AsyncSearch 异步搜索,针对超长时间查询
func (s *Store) AsyncSearch(body string) (string, error) {
	res, err := s.es.AsyncSearch.Submit(
		s.es.AsyncSearch.Submit.WithContext(context.Background()),
		s.es.AsyncSearch.Submit.WithIndex(s.IndexName),
		s.es.AsyncSearch.Submit.WithBody(strings.NewReader(body)),
		s.es.AsyncSearch.Submit.WithKeepOnCompletion(true),                        // 保持搜索结果
		s.es.AsyncSearch.Submit.WithWaitForCompletionTimeout(50*time.Millisecond), // 初始等待时间
		s.es.AsyncSearch.Submit.WithKeepAlive(10*time.Minute),                     // 任务有效时长
	)
	if err != nil {
		return "", fmt.Errorf("es error AsyncSearch: %s", err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("failed to close stmt: %v", err)
		}
	}(res.Body)

	if res.IsError() {
		return "", fmt.Errorf("es error : [%s] [%s] %s", res.Status(), s.IndexName, res.String())
	}

	// 提取异步搜索任务 ID
	var responseMap map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&responseMap); err != nil {
		return "", err
	}
	return responseMap["id"].(string), nil
}

// GetAsyncSearchResult 获取异步搜索结果
func (s *Store) GetAsyncSearchResult(taskID string) (*esapi.Response, error) {
	res, err := s.es.AsyncSearch.Get(
		taskID,
		s.es.AsyncSearch.Get.WithContext(context.Background()),
		s.es.AsyncSearch.Get.WithWaitForCompletionTimeout(1*time.Second), // 等待一段时间以查看任务是否完成
	)
	if err != nil {
		return nil, fmt.Errorf("es error GetAsyncSearchResult: %s", err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("failed to close stmt: %v", err)
		}
	}(res.Body)

	if res.IsError() {
		return nil, fmt.Errorf("es error  GetAsyncSearchResult: [%s] [%s] %s", res.Status(), s.IndexName, taskID)
	}

	// 输出返回的响应体
	return res, nil
}

// CancelAsyncSearch 取消异步搜索
func (s *Store) CancelAsyncSearch(taskID string) (bool, error) {
	res, err := s.es.AsyncSearch.Delete(
		taskID,
		s.es.AsyncSearch.Delete.WithContext(context.Background()),
	)
	if err != nil {
		return false, fmt.Errorf("es error CancelAsyncSearch: %s", err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("failed to close stmt: %v", err)
		}
	}(res.Body)

	if res.IsError() {
		return false, fmt.Errorf("es error CancelAsyncSearch: [%s] [%s] %s", res.Status(), s.IndexName, taskID)
	}

	// 输出返回的响应体
	return true, nil
}
