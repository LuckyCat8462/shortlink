package types

type CreateLinkReq struct {
	LongUrl    string `json:"longUrl"`
	SceneId    string `json:"sceneId,optional"`
	ExpireDays *int   `json:"expireDays,optional"`
}

type CreateLinkResp struct {
	ShortUrl  string  `json:"shortUrl"`
	ShortCode string  `json:"shortCode"`
	LongUrl   string  `json:"longUrl"`
	ExpireAt  *string `json:"expireAt"`
	Created   bool    `json:"created"`
}

type GetLinkReq struct {
	ShortCode string `path:"shortCode"`
}

type GetLinkResp struct {
	ShortCode string  `json:"shortCode"`
	LongUrl   string  `json:"longUrl"`
	SceneId   string  `json:"sceneId"`
	Status    int     `json:"status"`
	ExpireAt  *string `json:"expireAt"`
	CreatedAt string  `json:"createdAt"`
}

type GetStatsReq struct {
	ShortCode string `path:"shortCode"`
}

type GetStatsResp struct {
	ShortCode   string  `json:"shortCode"`
	Pv          int64   `json:"pv"`
	Uv          int64   `json:"uv"`
	LastClickAt *string `json:"lastClickAt"`
}

type RedirectReq struct {
	ShortCode string `path:"shortCode"`
}
