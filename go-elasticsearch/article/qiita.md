# [Golang]go-elasticsearchでレスポンスをmockする方法

## 本記事でやること
- [go-elasticsearch](https://github.com/elastic/go-elasticsearch)を使ってElasticsearchからのレスポンスをmockする
- レスポンスをmockした単体テストを書く
- 内部実装のコードリーディングを通して何をmockしているのかを理解する


## 対象読者

- これから[go-elasticsearch](https://github.com/elastic/go-elasticsearch)を使って実装を始める方

## 使用言語

- Go 1.21.0

## 実装

### 前提
Elasticsearchにあるインデックスに対して検索クエリを実行し結果を返す`Search`メソッドを実装しました。
今回は、この`Search`メソッドが返すレスポンスをmockします。

Searchメソッドは、検索対象のインデックス名と実行するクエリを文字列として受け取り、検索結果をバイト配列で返します。
また、Searchメソッドは`esapi.SearchRequest`構造体が持つ`Do`メソッドをラップすることで、Elasticsearchに対して検索リクエストを送信しています。


```golang
type esHandler struct {
	client *elasticsearch.Client
}

func NewEsHandler(client *elasticsearch.Client) *esHandler {
	return &esHandler{
		client: client,
	}
}

func (e *esHandler) Search(ctx context.Context, index, query string) ([]byte, error) {
	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  strings.NewReader(query),
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e errResponse
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("faild to read err response body: %w", err)
		}
		if err := json.Unmarshal(body, &e); err != nil {
			return nil, fmt.Errorf("faild to unmarsal err response body: %w", err)
		}
		return nil, fmt.Errorf("failt to search: [%d] %s", e.Status, e.Error.Cause[0].Reason)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

type errResponse struct {
	Status int      `json:"status"`
	Error  errCause `json:"error"`
}

type errCause struct {
	Cause []errReason `json:"root_cause"`
}

type errReason struct {
	Reason string `json:"reason"`
}
```

### mockの方法

簡単なテストコードを通して、`Search`メソッドのレスポンスをmockします。
実装は以下の通りです。

mockするために行っていることは単純で`elasticsearch.Config`構造体の`Transport`フィールドに`http.RoundTripper`型を満たす構造体(`mockTransport`)を渡すだけです。
この`mockTransport`構造体は、`RoundTrip`メソッドを実装しており、`http.RoundTripper`インターフェースを満たしています。このRoundTripメソッドでは、mockTransport構造体が持つ、http.Responseを返すようにしています。

つまり、`http.RoundTripper`インターフェースが持つ、`RoundTrip`メソッドが返す値(`http.Response`型)をmockすることで、Elasitcsearchからのレスポンスをmockすることができます。

```golang
type mockTransport struct {
response *http.Response
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
return m.response, nil
}

func TestSearch(t *testing.T) {
    body := `{
     "took" : 64,
     "timed_out" : false,
     "_shards" : {
        "total" : 1,
        "successful" : 1,
        "skipped" : 0,
        "failed" : 0
     },
     "hits" : {
        "total" : {
          "value" : 1,
          "relation" : "eq"
        },
        "max_score" : 1.0,
        "hits" : [
          {
            "_index" : "test_index",
            "_id" : "1",
            "_score" : 1.0,
            "_source" : {
              "title" : "hogehoge"
            }
          }
        ]
     }
    }`
	mockTrans := &mockTransport{
	    response: &http.Response{
	        StatusCode: http.StatusOK,
	        Body:       io.NopCloser(strings.NewReader(body)),
	        Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
	    },
	}

	cfg := elasticsearch.Config{
	    Transport: mockTrans,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
	    t.Fatal(err)
	}
	es := NewEsHandler(client)
	res, err := es.Search(context.Background(), "test", "")
	if err != nil {
        t.Fatal(err)
    }
    var r esResponse
    if err := json.Unmarshal(res, &r); err != nil {
        t.Fatal(err)
    }

    if r.Hits.Hits[0].Source.Title != "hogehoge" {
        t.Errorf("unexpected result: %s", r.Hits.Hits[0].Source.Title)
    }
}
```


## コードリーディング

:::note warn
以下ではgo-elasticsearchライブラリーの内部実装を読み進めていくため、記事の内容が少し長くなります。気になる方だけ読んでください。
:::

なぜRoundTripメソッドをmockすることでレスポンスをmockできるのかをgo-elasticのコードリーディングを通して理解します。


### 結論

先に結論だけを順序立てて書きます。

1. `Do`メソッドの内部で実行されている`esapi.Transport`インターフェースの`Perform`メソッドの具象は`elasticsearch.BaseClient`構造体の`Perform`メソッドである
2. `elasticsearch.BaseClient`構造体が持つ`Peform`メソッドの内部で実行されている`elastictransport.Interface`インターフェースの`Perform`メソッドの具象は`elastictransport.Client`構造体の`Perform`メソッドである
3. `elastictransport.Interface`は[elastic-transport-go](https://github.com/elastic/elastic-transport-go)ライブラリーで定義されており、Elasticsearchとデータの送受信を行うためのインターフェースを提供している
4. `elastictransport.Client`構造体が持つ`Perform`メソッドの内部で実行されているのは`http.RoundTripper`インターフェースの`RoundTrip`メソッドである
5. `RoundTrip`メソッドの具象は、ユーザーが`elasticsearch.Config`で定義した`http.RoundTripper`インターフェースの実装を満たす構造体の`RoundTrip`メソッドである

### `Do`メソッドの内部処理

https://github.com/elastic/go-elasticsearch/blob/main/esapi/api.search.go#L119

先で実装した`Search`メソッドは、`esapi.SearchRequest`構造体が持つ`Do`メソッドをラップしています。

まず、`Do`メソッドの実装を見てみます。以下に`Do`メソッドの実装を抜粋します。


`Do`メソッドの[L389](https://github.com/elastic/go-elasticsearch/blob/main/esapi/api.search.go#L389)でレスポンスを受け取り、[L400-L404](https://github.com/elastic/go-elasticsearch/blob/main/esapi/api.search.go#L400-L404)で`Response`構造体に詰め替えていることがわかります。
つまり、[L389](https://github.com/elastic/go-elasticsearch/blob/main/esapi/api.search.go#L389)で実行されている`Transport`型の`Perform`メソッドからElasticsearchからのレスポンスが渡って来ていることがわかります。

```golang:api.search.go
func (r SearchRequest) Do(providedCtx context.Context, transport Transport) (*Response, error) {

    ...省略
    // Elasticsearchからのレスポンスが渡ってきている
    res, err := transport.Perform(req)

    ...省略
	response := Response{
		StatusCode: res.StatusCode,
		Body:       res.Body,
		Header:     res.Header,
	}

	return &response, nil
```

### `Transport`型の`Perform`メソッドの具象

https://github.com/elastic/go-elasticsearch/blob/main/esapi/esapi.go#L33-L35

次に、`Transport`型の`Perform`メソッドの実装がどこで定義されているのかを探します。
`Do`メソッドが呼ばれている`Search`メソッドに戻ります。以下のコードから`Do`メソッドの`transport`引数(Transport`型)には`elasticsearch.Client`構造体が渡されていることがわかります。
つまり、`elasticsearch.Client`構造体が`Transport`インターフェイスの実装を満たすように`Perform`メソッドを持っていることがわかります。


```golang:api.search.go

type esHandler struct {
    client *elasticsearch.Client
}

func NewEsHandler(client *elasticsearch.Client) *esHandler {
    return &esHandler{
        client: client,
    }
}

func (e *esHandler) Search(ctx context.Context, index, query string) ([]byte, error) {
    req := esapi.SearchRequest{
        Index: []string{index},
        Body:  strings.NewReader(query),
    }

    res, err := req.Do(ctx, e.client)
```

`elasticsearch.Client`構造体が持つ`Peform`メソッドを探します。
まず、`elasticsearch.Client`構造体が定義されている`elasticsearch`パッケージの[elasticsearch.go](https://github.com/elastic/go-elasticsearch/blob/main/elasticsearch.go)をみてみます。
`elasticsearch.Client`構造体の`Perform`メソッドは確認できませんでしたが、その代わり`BaseClient`構造体が持つ`Perform`メソッドを確認することができました。([L322](https://github.com/elastic/go-elasticsearch/blob/main/elasticsearch.go#L322))以下に実装を抜粋します。

そして、`BaseClient`構造体は`Client`構造体に埋め込まれていることがわかります。つまり、`Client`構造体は`BaseClient`構造体の埋め込みによって`Transport`インターフェイスの実装を満たしています。

```golang:elasticsearch.go

つまり、`Do`メソッドのなかで実行されていた`tansport.Perform`メソッドの具象は以下で定義されている`BaseClient`構造体の`Perform`メソッドであることがわかりました。

```golang:elasticsearch.go

...省略

// Client represents the Functional Options API.
type Client struct {
	BaseClient
	*esapi.API
}

...省略

// Perform delegates to Transport to execute a request and return a response.
func (c *BaseClient) Perform(req *http.Request) (*http.Response, error) {
    ...省略
	// Retrieve the original request.
	res, err := c.Transport.Perform(req)
	...省略
}

```

### `elastictransport.Interface`の`Perform`メソッドの具象

https://github.com/elastic/elastic-transport-go/blob/main/elastictransport/elastictransport.go#L50-L52

`BaseClient`構造体の`Perform`メソッドの中でさらに`Peform`メソッドが呼ばれていることがわかります。
これは`BaseClient`構造体の`Transport`フィールド(`elasticsearch.Interface`型)が持つ、`Perform`メソッドを呼び出しています。

まず、`BaseClient`構造体の`Transport`フィールドがどのように定義されているのかを探すために`BaseClient`構造体がイニシャライズされる`NewClient`関数の実装を見てみます。

`NewClient`関数の[L189](https://github.com/elastic/go-elasticsearch/blob/main/elasticsearch.go#L189)で定義されている`Transport`フィールド(`elastictransport.Interface`型)の値は`newTransport`関数で初期化されています。([L179](https://github.com/elastic/go-elasticsearch/blob/main/elasticsearch.go#L179))

以下に`NewClient`関数の実装を抜粋します。

```golang:elasticsearch.go

...省略

// BaseClient represents the Elasticsearch client.
type BaseClient struct {
	Transport           elastictransport.Interface
	metaHeader          string
	compatibilityHeader bool

	disableMetaHeader   bool
	productCheckMu      sync.RWMutex
	productCheckSuccess bool
}

...省略

func NewClient(cfg Config) (*Client, error) {
	tp, err := newTransport(cfg)
	if err != nil {
		return nil, err
	}

	compatHeaderEnv := os.Getenv(esCompatHeader)
	compatibilityHeader, _ := strconv.ParseBool(compatHeaderEnv)

	client := &Client{
		BaseClient: BaseClient{
			Transport:           tp,
			disableMetaHeader:   cfg.DisableMetaHeader,
			metaHeader:          initMetaHeader(tp),
			compatibilityHeader: cfg.EnableCompatibilityMode || compatibilityHeader,
		},
	}
	client.API = esapi.New(client)

	if cfg.DiscoverNodesOnStart {
		go client.DiscoverNodes()
	}

	return client, nil
}
```

`elastictransport.Interface`インターフェースは[elastic-transport-go](https://github.com/elastic/elastic-transport-go/tree/main)ライブラリーで定義されています。このライブラリーはREADMEにあるように`go-elasticsearch`で使用されるトランスポートインターフェースを提供しています。
つまり、Elasticsearchとデータの送受信を行うためのインターフェースを提供していることがわかります。

https://github.com/elastic/elastic-transport-go/blob/main/README.md

>It provides the Transport interface used by go-elasticsearch, connection pool, cluster discovery, and multiple loggers.


`newTransport`関数の実装に話を戻すと、`newTransport`関数はElasticsearchとの通信を行うためのHTTPクライアントである`elastictransport.Client`構造体を初期化しています。
そして、`elastictransport.Client`構造体は`elastictransport.Interface`インターフェースの実装を満たしているので、`Perform`メソッドを持っていることがわかります。

以下に`newTransport`関数の実装を抜粋します。

```golang:elasticsearch.go

...省略

func newTransport(cfg Config) (*elastictransport.Client, error) {
    ...省略
    tp, err := elastictransport.New(tpConfig)
    ...省略
    return tp, nil
}
```


`elastictransport.Client`構造体の`Perform`メソッドの実装を見てみます。

以下に`Perform`メソッドの実装を抜粋します。


やっとここで、`RoundTrip`メソッドが出てきました。この`RoundTrip`メソッドは`elastictransport.Client`構造体がもつ`transport`フィールド(`http.RoundTripper`型)の`RoundTrip`メソッドを呼び出しています。
また、`elastictransport.Client`構造体の`transport`フィールドは、`elastictransport.Config`構造体の`Transport`フィールドの値から渡ってきていることがわかります。

```golang:elastictransport.go

func New(cfg Config) (*Client, error) {
    ...省略
    client := Client{
        ...
        transport: cfg.Transport,
        ...
    }
    ...
    return &client, nil
}

// Perform executes the request and returns a response or error.
func (c *Client) Perform(req *http.Request) (*http.Response, error) {

    ...省略
    res, err = c.transport.RoundTrip(req)
    ...省略
    return res, err
```

`elastictransport.Config`構造体の`Transport`フィールドは、`NewClinet`関数の`cfg`引数(`elasitcsearch.Config`型)の`Transport`フィールドの値から渡ってきています。

ここでやっと、ユーザーが入力した`http.RoundTripper`インターフェースの実装を満たす構造体が影響してくることがわかります。

よって、`http.RoundTripper`インターフェースの実装を満たす構造体を`elasticsearch.Config`構造体の`Transport`フィールドに渡すことで、Elasticsearchとデータの送受信を行う`elastictransport.Client`構造体の`RoundTrip`メソッドをmockすることができます。
この`RoundTrip`メソッドはHTTP通信の実態を担っているため、このメソッドをmockすることでElasticsearchからのレスポンスをmockすることができるということがわかります。

## まとめ
今回は、[go-elasticsearch](https://github.com/elastic/go-elasticsearch/tree/main)ライブラリーを使ってElasticsearchからのレスポンスをmockする方法を紹介しました。
また、内部実装のコードリーディングを通して、なぜ`RoundTrip`メソッドをmockすることでレスポンスをmockできるのかを順を追って説明しました。

この記事が[go-elasticsearch](https://github.com/elastic/go-elasticsearch/tree/main)の内部実装を理解する手助けになれば幸いです。
