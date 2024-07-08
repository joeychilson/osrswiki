package osrswiki

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	pricesEndpoint = "https://prices.runescape.wiki/api/v1"
)

type Client struct {
	userAgent  string
	httpClient *http.Client
}

func NewClient(userAgent string) *Client {
	return &Client{
		userAgent:  userAgent,
		httpClient: &http.Client{},
	}
}

type World string

const (
	WorldRegular    World = "osrs"
	WorldDeadman    World = "dmm"
	WorldFreshStart World = "fsw"
)

type LatestPrice struct {
	High     int64 `json:"high"`
	HighTime int64 `json:"highTime"`
	Low      int64 `json:"low"`
	LowTime  int64 `json:"lowTime"`
}

func (c *Client) LatestPrices(ctx context.Context, world World, itemIDs ...int16) (map[int16]LatestPrice, error) {
	url := fmt.Sprintf("%s/%s/latest", pricesEndpoint, world)

	query := make(map[string]string)
	if len(itemIDs) > 0 {
		ids := make([]string, len(itemIDs))
		for i, id := range itemIDs {
			ids[i] = strconv.FormatInt(int64(id), 10)
		}
		query["id"] = strings.Join(ids, ",")
	}

	body, err := c.doRequest(ctx, url, query)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data map[string]LatestPrice `json:"data"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	result := make(map[int16]LatestPrice)
	for itemIDStr, data := range response.Data {
		itemID, err := strconv.ParseInt(itemIDStr, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("error parsing item ID: %w", err)
		}
		result[int16(itemID)] = data
	}
	return result, nil
}

type ItemMapping struct {
	ID       int    `json:"id"`
	Icon     string `json:"icon"`
	Name     string `json:"name"`
	Examine  string `json:"examine"`
	Members  bool   `json:"members"`
	Value    int    `json:"value"`
	HighAlch int    `json:"highalch"`
	LowAlch  int    `json:"lowalch"`
	Limit    int    `json:"limit"`
}

func (c *Client) ItemMapping(ctx context.Context, world World) ([]ItemMapping, error) {
	url := fmt.Sprintf("%s/%s/%s", pricesEndpoint, world, "mapping")

	body, err := c.doRequest(ctx, url, nil)
	if err != nil {
		return nil, err
	}

	var items []ItemMapping
	err = json.Unmarshal(body, &items)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return items, nil
}

type TimeInterval string

const (
	FiveMinutes     TimeInterval = "5m"
	OneHour         TimeInterval = "1h"
	SixHours        TimeInterval = "6h"
	TwentyFourHours TimeInterval = "24h"
)

type PriceData struct {
	AvgHighPrice    int64 `json:"avgHighPrice"`
	HighPriceVolume int64 `json:"highPriceVolume"`
	AvgLowPrice     int64 `json:"avgLowPrice"`
	LowPriceVolume  int64 `json:"lowPriceVolume"`
}

func (c *Client) PriceData(ctx context.Context, world World, interval TimeInterval, timestamp *time.Time) (map[int16]PriceData, error) {
	if interval != FiveMinutes && interval != OneHour {
		return nil, fmt.Errorf("only 5m and 1h intervals are supported")
	}

	url := fmt.Sprintf("%s/%s/%s", pricesEndpoint, world, interval)

	query := make(map[string]string)
	if timestamp != nil {
		query["timestamp"] = fmt.Sprintf("%d", timestamp.Unix())
	}

	body, err := c.doRequest(ctx, url, query)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data map[string]PriceData `json:"data"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	result := make(map[int16]PriceData)
	for itemIDStr, data := range response.Data {
		itemID, err := strconv.ParseInt(itemIDStr, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("error parsing item ID: %w", err)
		}
		result[int16(itemID)] = data
	}
	return result, nil
}

type TimeseriesData struct {
	Timestamp       int64 `json:"timestamp"`
	AvgHighPrice    int64 `json:"avgHighPrice"`
	AvgLowPrice     int64 `json:"avgLowPrice"`
	HighPriceVolume int64 `json:"highPriceVolume"`
	LowPriceVolume  int64 `json:"lowPriceVolume"`
}

func (c *Client) Timeseries(ctx context.Context, world World, timestep TimeInterval, itemID int16) ([]TimeseriesData, error) {
	url := fmt.Sprintf("%s/%s/timeseries", pricesEndpoint, world)

	query := map[string]string{
		"id":       fmt.Sprintf("%d", itemID),
		"timestep": string(timestep),
	}

	body, err := c.doRequest(ctx, url, query)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data []TimeseriesData `json:"data"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return response.Data, nil
}

func (c *Client) doRequest(ctx context.Context, url string, query map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	q := req.URL.Query()
	for key, value := range query {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return body, nil
}
