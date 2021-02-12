package telegram

type Response struct {
	OK         bool                   `json:"ok"`
	Desription string                 `json:"description,omitempty"`
	Result     interface{}            `json:"result,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

type GetMeResponse struct {
	ID                      int    `json:"id" example:"1609917215"`
	IsBot                   bool   `json:"is_bot" example:"true"`
	FirstName               string `json:"first_name" example:"ShopFlow"`
	UserName                string `json:"username" example:"ShopFlowBot"`
	CanJoinGroups           bool   `json:"can_join_groups" example:"true"`
	CanReadAllGroupMessages bool   `json:"can_read_all_group_messages" example:"false"`
	SupportsInlineQueries   bool   `json:"supports_inline_queries" example:"false"`
}

type SetWebHookRequest struct {
	URL                string   `json:"url" doc:"HTTPS url to send updates to. Use an empty string to remove webhook integration"`
	Certificate        string   `json:"certificate,omitempty" doc:"Upload your public key certificate so that the root certificate in use can be checked. See our self-signed guide for details."`
	IP                 string   `json:"ip_address,omitempty" doc:"The fixed IP address which will be used to send webhook requests instead of the IP address resolved through DNS"`
	MaxConn            int      `json:"max_connections,omitempty" doc:"Maximum allowed number of simultaneous HTTPS connections to the webhook for update delivery, 1-100. Defaults to 40. Use lower values to limit the load on your bot's server, and higher values to increase your bot's throughput."`
	AllowedUpdates     []string `json:"allowed_updates,omitempty" doc:"A JSON-serialized list of the update types you want your bot to receive. For example, specify [“message”, “edited_channel_post”, “callback_query”] to only receive updates of these types. See Update for a complete list of available update types. Specify an empty list to receive all updates regardless of type (default). If not specified, the previous setting will be used. Please note that this parameter doesn't affect updates created before the call to the setWebhook, so unwanted updates may be received for a short period of time."`
	DropPendingUpdates bool     `json:"drop_pending_updates,omitempty" doc:"Pass True to drop all pending updates"`
}

type SetWebHookResponse bool
