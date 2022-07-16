package twitter

type TwitterClient struct {
	twitterOauthClient *TwitterOauthClient
}

type TwitterClientArgs struct {
}

func NewTwitterClient(args *TwitterClientArgs) *TwitterClient {
	c := new(TwitterClient)
	c.twitterOauthClient = NewTwitterOauthClient()
	return c
}

// GetProfileInformation: Obtener información de perfil
func (t *TwitterClient) GetProfileInformation(profileId string) (*ProfileInformation, error) {
	return t.twitterOauthClient.GetProfileInformation(profileId)
}

// GetLatestNTweetsForProfile: Obtenga los N Tweets más recientes para el perfil
func (t *TwitterClient) GetLatestNTweetsForProfile(profileId string, elements int64) (*[]LatestNTweetsForProfile, error) {
	return t.twitterOauthClient.GetLatestNTweetsForProfile(profileId, elements)
}

// ModifyProfileInformation: Modificar la información del perfil
func (t *TwitterClient) ModifyProfileInformation(profileId string, payload interface{}) interface{} {
	return new(map[string]interface{})
}

// GetProfileRequests: Obtener solicitudes de perfil
func (t *TwitterClient) GetProfileRequests(profileId string) interface{} {
	return new(map[string]interface{})
}

// Close
func (t *TwitterClient) Close() {

}
