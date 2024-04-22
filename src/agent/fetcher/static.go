package fetcher

import "encoding/json"

type StaticFetcher struct {
	cfg StaticFetcherConfig
}

type StaticFetcherConfig struct {
	Filepath string `json:"file_path"`
}

func (fetcher *StaticFetcher) Name() string {
	return "static"
}

func (fetcher *StaticFetcher) Setup(cfg json.RawMessage) error {
	var fetcherConfig = StaticFetcherConfig{}
	err := json.Unmarshal(cfg, &fetcherConfig)
	if err != nil {
		return err
	}

	fetcher.cfg = fetcherConfig
	return nil
}

func (fetcher *StaticFetcher) Fetch() (*FetchResult, error) {
	//TODO implement me
	panic("implement me")
}

func init() {
	RegisterFetcher(&StaticFetcher{})
}
