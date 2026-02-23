package es

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"shortvideo/pkg/config"
	"sync"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

var (
	esInstance *ESManager
	esOnce     sync.Once
)

type ESManager struct {
	client *elasticsearch.Client
}

type IndexMapping struct {
	Settings IndexSettings     `json:"settings"`
	Mappings PropertiesMapping `json:"mappings,omitempty"`
}

type IndexSettings struct {
	NumberOfShards   int `json:"number_of_shards,omitempty"`
	NumberOfReplicas int `json:"number_of_replicas,omitempty"`
}

type PropertiesMapping struct {
	Properties map[string]PropertyMapping `json:"properties,omitempty"`
}

type PropertyMapping struct {
	Type     string                 `json:"type,omitempty"`
	Analyzer string                 `json:"analyzer,omitempty"`
	Format   string                 `json:"format,omitempty"`
	Fields   map[string]interface{} `json:"fields,omitempty"`
}

type SearchQuery struct {
	Query interface{}              `json:"query,omitempty"`
	From  int                      `json:"from,omitempty"`
	Size  int                      `json:"size,omitempty"`
	Sort  []map[string]interface{} `json:"sort,omitempty"`
}

type SearchResult struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Hits     struct {
		Total struct {
			Value    int64  `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		Hits []json.RawMessage `json:"hits"`
	} `json:"hits"`
}

// 创建ES管理器
func NewESManager() (*ESManager, error) {
	var err error
	esOnce.Do(func() {
		cfg := config.Get().Elasticsearch
		esCfg := elasticsearch.Config{
			Addresses: []string{cfg.URL},
			Username:  cfg.Username,
			Password:  cfg.Password,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
		client, clientErr := elasticsearch.NewClient(esCfg)
		if clientErr != nil {
			log.Printf("创建ES客户端失败: %v", clientErr)
			err = clientErr
			return
		}
		_, infoErr := client.Info()
		if infoErr != nil {
			log.Printf("ES连接测试失败: %v", infoErr)
		}
		esInstance = &ESManager{
			client: client,
		}
		log.Println("ES客户端初始化成功")
	})
	return esInstance, err
}

// 获取ES客户端
func GetESClient() (*ESManager, error) {
	if esInstance == nil {
		return NewESManager()
	}
	return esInstance, nil
}

// 创建索引
func (es *ESManager) CreateIndex(indexName string, mapping IndexMapping) error {
	exists, err := es.IndexExists(indexName)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("索引 %s 已存在", indexName)
		return nil
	}

	body, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("序列化映射失败: %w", err)
	}

	resp, err := es.client.Indices.Create(
		indexName,
		es.client.Indices.Create.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		// 读取并打印详细的错误信息
		var errorBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorBody); err != nil {
			log.Printf("解析错误响应失败: %v", err)
		} else {
			log.Printf("索引创建失败详情: %v", errorBody)
		}
		return fmt.Errorf("创建索引失败: %s", resp.Status())
	}

	log.Printf("索引 %s 创建成功", indexName)
	return nil
}

// 删除索引
func (es *ESManager) DeleteIndex(indexName string) error {
	exists, err := es.IndexExists(indexName)
	if err != nil {
		return err
	}
	if !exists {
		log.Printf("索引 %s 不存在", indexName)
		return nil
	}

	resp, err := es.client.Indices.Delete(
		[]string{indexName},
	)
	if err != nil {
		return fmt.Errorf("删除索引失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("删除索引失败: %s", resp.Status())
	}

	log.Printf("索引 %s 删除成功", indexName)
	return nil
}

// 检查索引是否存在
func (es *ESManager) IndexExists(indexName string) (bool, error) {
	resp, err := es.client.Indices.Exists(
		[]string{indexName},
	)
	if err != nil {
		return false, fmt.Errorf("检查索引存在性失败: %w", err)
	}
	defer resp.Body.Close()

	return !resp.IsError(), nil
}

// 添加文档
func (es *ESManager) AddDocument(indexName string, id string, document interface{}) error {
	body, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("序列化文档失败: %w", err)
	}

	var resp *esapi.Response
	if id != "" {
		resp, err = es.client.Index(
			indexName,
			bytes.NewReader(body),
			es.client.Index.WithDocumentID(id),
			es.client.Index.WithRefresh("true"),
		)
	} else {
		resp, err = es.client.Index(
			indexName,
			bytes.NewReader(body),
			es.client.Index.WithRefresh("true"),
		)
	}

	if err != nil {
		return fmt.Errorf("添加文档失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("添加文档失败: %s", resp.Status())
	}

	return nil
}

// 更新文档
func (es *ESManager) UpdateDocument(indexName string, id string, document interface{}) error {
	updateBody := map[string]interface{}{
		"doc": document,
	}

	body, err := json.Marshal(updateBody)
	if err != nil {
		return fmt.Errorf("序列化更新内容失败: %w", err)
	}

	resp, err := es.client.Update(
		indexName,
		id,
		bytes.NewReader(body),
		es.client.Update.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("更新文档失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("更新文档失败: %s", resp.Status())
	}

	return nil
}

// 删除文档
func (es *ESManager) DeleteDocument(indexName string, id string) error {
	resp, err := es.client.Delete(
		indexName,
		id,
		es.client.Delete.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("删除文档失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("删除文档失败: %s", resp.Status())
	}

	return nil
}

// 搜索文档
func (es *ESManager) Search(indexName string, query SearchQuery, result interface{}) error {
	body, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("序列化查询失败: %w", err)
	}

	resp, err := es.client.Search(
		es.client.Search.WithIndex(indexName),
		es.client.Search.WithBody(bytes.NewReader(body)),
		es.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return fmt.Errorf("搜索失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("搜索失败: %s", resp.Status())
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("解析搜索结果失败: %w", err)
	}

	return nil
}

// 健康检查
func (es *ESManager) HealthCheck() error {
	resp, err := es.client.Cluster.Health()
	if err != nil {
		return fmt.Errorf("健康检查失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("健康检查失败: %s", resp.Status())
	}

	log.Println("ES健康检查正常")
	return nil
}

// 生成视频索引映射
func GenerateVideoMapping() IndexMapping {
	return IndexMapping{
		Settings: IndexSettings{
			NumberOfShards:   1,
			NumberOfReplicas: 0,
		},
		Mappings: PropertiesMapping{
			Properties: map[string]PropertyMapping{
				"id": {
					Type: "long",
				},
				"user_id": {
					Type: "long",
				},
				"title": {
					Type: "text",
					Fields: map[string]interface{}{
						"keyword": map[string]interface{}{
							"type":         "keyword",
							"ignore_above": 256,
						},
					},
				},
				"description": {
					Type: "text",
				},
				"cover_url": {
					Type: "keyword",
				},
				"video_url": {
					Type: "keyword",
				},
				"tags": {
					Type: "keyword",
				},
				"view_count": {
					Type: "long",
				},
				"like_count": {
					Type: "long",
				},
				"comment_count": {
					Type: "long",
				},
				"share_count": {
					Type: "long",
				},
				"created_at": {
					Type:   "date",
					Format: "yyyy-MM-dd HH:mm:ss",
				},
			},
		},
	}
}

// 生成用户索引映射
func GenerateUserMapping() IndexMapping {
	return IndexMapping{
		Settings: IndexSettings{
			NumberOfShards:   1,
			NumberOfReplicas: 0,
		},
		Mappings: PropertiesMapping{
			Properties: map[string]PropertyMapping{
				"id": {
					Type: "long",
				},
				"username": {
					Type: "keyword",
				},
				"avatar": {
					Type: "keyword",
				},
				"about": {
					Type: "text",
				},
				"follow_count": {
					Type: "long",
				},
				"follower_count": {
					Type: "long",
				},
				"created_at": {
					Type:   "date",
					Format: "yyyy-MM-dd HH:mm:ss",
				},
			},
		},
	}
}

// 生成直播索引映射
func GenerateLiveMapping() IndexMapping {
	return IndexMapping{
		Settings: IndexSettings{
			NumberOfShards:   1,
			NumberOfReplicas: 0,
		},
		Mappings: PropertiesMapping{
			Properties: map[string]PropertyMapping{
				"id": {
					Type: "long",
				},
				"host_id": {
					Type: "long",
				},
				"title": {
					Type: "text",
					Fields: map[string]interface{}{
						"keyword": map[string]interface{}{
							"type":         "keyword",
							"ignore_above": 256,
						},
					},
				},
				"cover_url": {
					Type: "keyword",
				},
				"viewer_count": {
					Type: "long",
				},
				"is_live": {
					Type: "boolean",
				},
				"created_at": {
					Type:   "date",
					Format: "yyyy-MM-dd HH:mm:ss",
				},
			},
		},
	}
}
