package whop

type AuthTokenRequest struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectUrl  string `json:"redirect_uri"`
}

type WhopErrorResponse struct {
	Error struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	} `json:"error"`
}

type WhopValidateLicenseResponse struct {
	ID                string `json:"id,omitempty"`
	Product           string `json:"product,omitempty"`
	User              string `json:"user,omitempty"`
	Plan              string `json:"plan,omitempty"`
	Email             string `json:"email,omitempty"`
	Status            string `json:"status,omitempty"`
	Valid             bool   `json:"valid,omitempty"`
	CancelAtPeriodEnd bool   `json:"cancel_at_period_end,omitempty"`
	PaymentProcessor  string `json:"payment_processor,omitempty"`
	LicenseKey        string `json:"license_key,omitempty"`
	Metadata          struct {
		NewKey string `json:"newKey,omitempty"`
	} `json:"metadata,omitempty"`
	Quantity              int    `json:"quantity,omitempty"`
	WalletAddress         string `json:"wallet_address,omitempty"`
	CustomFieldsResponses struct {
	} `json:"custom_fields_responses,omitempty"`
	Discord struct {
		ID       string `json:"id,omitempty"`
		Username string `json:"username,omitempty"`
	} `json:"discord,omitempty"`
	NftTokens []struct {
		TokenId       string `json:"token_id,omitempty"`
		CurrentHolder string `json:"current_holder,omitempty"`
		SmartContract struct {
			ContractAddress string `json:"contract_address,omitempty"`
			ContractName    string `json:"contract_name,omitempty"`
		} `json:"smart_contract,omitempty"`
		Balance  int `json:"balance,omitempty"`
		Metadata struct {
		} `json:"metadata,omitempty"`
	} `json:"nft_tokens,omitempty"`
	ExpiresAt          int    `json:"expires_at,omitempty"`
	RenewalPeriodStart int    `json:"renewal_period_start,omitempty"`
	RenewalPeriodEnd   int    `json:"renewal_period_end,omitempty"`
	CreatedAt          int    `json:"created_at,omitempty"`
	ManageURL          string `json:"manage_url,omitempty"`
	AffiliatePageURL   string `json:"affiliate_page_url,omitempty"`
	AccessPass         string `json:"access_pass,omitempty"`
}
