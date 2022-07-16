package twitter

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/juliotorresmoreno/zemona/config"
)

type TwitterAuthorization struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type TwitterOauthClient struct {
	authorization *TwitterAuthorization
	apiKey        string
	apiKeySecret  string
}

func NewTwitterOauthClient() *TwitterOauthClient {
	c := &TwitterOauthClient{}
	config := config.GetConfig()
	c.apiKey = config.TwitterApiKey
	c.apiKeySecret = config.TwitterApiKeySecret

	return c
}

const (
	AUTH_URL                        = "https://api.twitter.com/oauth2/token"
	GET_PROFILE_INFORMATION         = "https://api.twitter.com/2/users/by"
	GET_LATEST_N_TWEETS_FOR_PROFILE = "https://api.twitter.com/2/tweets/search/recent"
)

func (t *TwitterOauthClient) basicAuth() string {
	str := t.apiKey + ":" + t.apiKeySecret
	sEnc := base64.StdEncoding.EncodeToString([]byte(str))
	return sEnc
}

func (t *TwitterOauthClient) DoAuth() error {
	payload := bytes.NewBufferString("grant_type=client_credentials")
	req, err := http.NewRequest("POST", AUTH_URL, payload)
	if err != nil {
		return err
	}
	basicAuthorization := t.basicAuth()
	req.Header.Add("Authorization", "Basic "+basicAuthorization)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Add("User-Agent", "OnnaSoft")
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	authorization := new(TwitterAuthorization)
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, authorization)
	if err != nil {
		return err
	}

	t.authorization = authorization
	return nil
}

type ProfileInformationResponse struct {
	Data []ProfileInformation `json:"data"`
}

type ProfileInformation struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

func (t *TwitterOauthClient) GetProfileInformation(profileId string) (*ProfileInformation, error) {
	profileInformationResponse := &ProfileInformationResponse{}
	rawUrl := fmt.Sprintf(
		"%v?usernames=%v&user.fields=created_at&expansions=pinned_tweet_id&tweet.fields=author_id,created_at",
		GET_PROFILE_INFORMATION,
		profileId,
	)
	u, _ := url.Parse(rawUrl)
	req, err := http.NewRequest("GET", u.String(), nil)

	if err != nil {
		return &ProfileInformation{}, err
	}
	if t.authorization == nil {
		t.DoAuth()
	}
	req.Header.Add("Authorization", "Bearer "+t.authorization.AccessToken)
	req.Header.Add("User-Agent", "OnnaSoft")
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return &ProfileInformation{}, err
	}

	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return &ProfileInformation{}, err
	}

	err = json.Unmarshal(content, profileInformationResponse)
	if err != nil {
		return &ProfileInformation{}, err
	}

	if len(profileInformationResponse.Data) == 0 {
		return &ProfileInformation{}, nil
	}

	return &profileInformationResponse.Data[0], nil
}

type LatestNTweetsForProfile struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	Text      string `json:"text"`
	AuthorId  string `json:"author_id"`
}

type LatestNTweetsForProfileResponse struct {
	Data *[]LatestNTweetsForProfile
}

func (t *TwitterOauthClient) GetLatestNTweetsForProfile(profileId string, elements int64) (*[]LatestNTweetsForProfile, error) {
	latestNTweetsForProfileResponse := &LatestNTweetsForProfileResponse{}
	rawUrl := fmt.Sprintf(
		"%v?query=from:%v&tweet.fields=created_at&expansions=author_id&user.fields=created_at&max_results=%v",
		GET_LATEST_N_TWEETS_FOR_PROFILE,
		profileId,
		elements,
	)
	u, _ := url.Parse(rawUrl)
	req, err := http.NewRequest("GET", u.String(), nil)

	if err != nil {
		return &[]LatestNTweetsForProfile{}, err
	}
	if t.authorization == nil {
		t.DoAuth()
	}
	req.Header.Add("Authorization", "Bearer "+t.authorization.AccessToken)
	req.Header.Add("User-Agent", "OnnaSoft")
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return &[]LatestNTweetsForProfile{}, err
	}

	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return &[]LatestNTweetsForProfile{}, err
	}
	os.Stdout.Write(content)
	os.Stdout.Write([]byte("\n"))

	err = json.Unmarshal(content, latestNTweetsForProfileResponse)
	if err != nil {
		return &[]LatestNTweetsForProfile{}, err
	}

	return latestNTweetsForProfileResponse.Data, nil
}
