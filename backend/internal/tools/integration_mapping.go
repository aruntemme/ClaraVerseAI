package tools

// ToolIntegrationMap maps tool names to their corresponding integration types.
// This enables automatic credential injection based on the tool being called.
// When a tool is executed, we look up its integration type and find a matching
// credential from the user's configured credentials.
var ToolIntegrationMap = map[string]string{
	// Communication tools
	"send_discord_message":     "discord",
	"send_slack_message":       "slack",
	"send_telegram_message":    "telegram",
	"send_teams_message":       "teams",
	"send_google_chat_message": "google_chat",
	"zoom_meeting":             "zoom",
	"twilio_send_sms":          "twilio",
	"twilio_send_whatsapp":     "twilio",
	"referralmonk_whatsapp":    "referralmonk",

	// Email tools
	"send_email":       "sendgrid",
	"send_brevo_email": "brevo",

	// Generic webhook
	"send_webhook": "custom_webhook",

	// REST API
	"api_request": "rest_api",

	// Notion tools
	"notion_search":         "notion",
	"notion_query_database": "notion",
	"notion_create_page":    "notion",
	"notion_update_page":    "notion",

	// GitHub tools
	"github_create_issue": "github",
	"github_list_issues":  "github",
	"github_get_repo":     "github",
	"github_add_comment":  "github",

	// GitLab tools
	"gitlab_projects": "gitlab",
	"gitlab_issues":   "gitlab",
	"gitlab_mrs":      "gitlab",

	// Linear tools
	"linear_issues":       "linear",
	"linear_create_issue": "linear",
	"linear_update_issue": "linear",

	// Jira tools
	"jira_issues":       "jira",
	"jira_create_issue": "jira",
	"jira_update_issue": "jira",

	// Productivity tools
	"clickup_tasks":       "clickup",
	"clickup_create_task": "clickup",
	"clickup_update_task": "clickup",
	"calendly_events":      "calendly",
	"calendly_event_types": "calendly",
	"calendly_invitees":    "calendly",

	// Airtable tools
	"airtable_list":   "airtable",
	"airtable_read":   "airtable",
	"airtable_create": "airtable",
	"airtable_update": "airtable",

	// Trello tools
	"trello_boards":      "trello",
	"trello_lists":       "trello",
	"trello_cards":       "trello",
	"trello_create_card": "trello",

	// CRM tools
	"leadsquared_leads":       "leadsquared",
	"leadsquared_create_lead": "leadsquared",
	"leadsquared_activities":  "leadsquared",
	"hubspot_contacts":        "hubspot",
	"hubspot_deals":           "hubspot",
	"hubspot_companies":       "hubspot",

	// Marketing tools
	"mailchimp_lists":          "mailchimp",
	"mailchimp_add_subscriber": "mailchimp",

	// Analytics tools
	"mixpanel_track":        "mixpanel",
	"mixpanel_user_profile": "mixpanel",
	"posthog_capture":       "posthog",
	"posthog_identify":      "posthog",
	"posthog_query":         "posthog",

	// E-commerce tools
	"shopify_products":  "shopify",
	"shopify_orders":    "shopify",
	"shopify_customers": "shopify",

	// Deployment tools
	"netlify_sites":         "netlify",
	"netlify_deploys":       "netlify",
	"netlify_trigger_build": "netlify",

	// Storage tools
	"s3_list":     "aws_s3",
	"s3_upload":   "aws_s3",
	"s3_download": "aws_s3",
	"s3_delete":   "aws_s3",

	// Social Media tools
	"x_search_posts":   "x_twitter",
	"x_post_tweet":     "x_twitter",
	"x_get_user":       "x_twitter",
	"x_get_user_posts": "x_twitter",

	// Database tools
	"mongodb_query": "mongodb",
	"mongodb_write": "mongodb",
	"redis_read":    "redis",
	"redis_write":   "redis",

	// Composio Google Sheets tools
	"googlesheets_read":          "composio_googlesheets",
	"googlesheets_write":         "composio_googlesheets",
	"googlesheets_append":        "composio_googlesheets",
	"googlesheets_create":        "composio_googlesheets",
	"googlesheets_get_info":      "composio_googlesheets",
	"googlesheets_list_sheets":   "composio_googlesheets",
	"googlesheets_search":        "composio_googlesheets",
	"googlesheets_clear":         "composio_googlesheets",
	"googlesheets_add_sheet":     "composio_googlesheets",
	"googlesheets_delete_sheet":  "composio_googlesheets",
	"googlesheets_find_replace":  "composio_googlesheets",
	"googlesheets_upsert_rows":   "composio_googlesheets",

	// Composio Gmail tools
	"gmail_send_email":     "composio_gmail",
	"gmail_fetch_emails":   "composio_gmail",
	"gmail_get_message":    "composio_gmail",
	"gmail_reply_to_thread": "composio_gmail",
	"gmail_create_draft":   "composio_gmail",
	"gmail_send_draft":     "composio_gmail",
	"gmail_list_drafts":    "composio_gmail",
	"gmail_add_label":      "composio_gmail",
	"gmail_list_labels":    "composio_gmail",
	"gmail_move_to_trash":  "composio_gmail",
}

// GetIntegrationTypeForTool returns the integration type for a given tool name.
// Returns empty string if the tool doesn't require credentials.
func GetIntegrationTypeForTool(toolName string) string {
	return ToolIntegrationMap[toolName]
}

// ToolRequiresCredential returns true if the tool requires a credential.
func ToolRequiresCredential(toolName string) bool {
	_, exists := ToolIntegrationMap[toolName]
	return exists
}
