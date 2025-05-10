package models

// Form labels
const (
	LabelURL      = "URL"
	LabelMethod   = "Method"
	LabelHeaders  = "Headers"
	LabelBody     = "Body"
	LabelResponse = "Response"
	LabelSend     = "Send"
	LabelForms    = "Forms"
	LabelForm     = "Form"
)

// Theme labels
const (
	ThemeLight = "Light"
	ThemeDark  = "Dark"
)

// Filter labels
const (
	FilterPlaceholder = "Filter by form name"
)

// Error messages
const (
	ErrReadingCollection = "error reading Postman collection: %v"
	ErrParsingCollection = "error parsing Postman collection: %v"
	ErrCreatingRequest   = "Error creating request: %v"
	ErrSendingRequest    = "Error sending request: %v"
	ErrReadingResponse   = "Error reading response: %v"
	ErrRequestCancelled  = "Запрос отменен"
	ErrRequestInProgress = "Запрос выполняется %s 🚀"
)

// Log messages
const (
	LogStartingApp        = "Starting application..."
	LogWindowReady        = "Application window created and ready"
	LogSettingForm        = "Setting form"
	LogLoadingForms       = "Error loading forms"
	LogLoadedVariables    = "Loaded Postman variables"
	LogTotalItems         = "Total items in collection"
	LogProcessingItem     = "Processing item"
	LogURLPath            = "URL Path"
	LogFormID             = "Form ID"
	LogAddedForm          = "Added form"
	LogSkippingItem       = "Skipping item: invalid URL path length"
	LogTotalForms         = "Total forms loaded"
	LogLoadedForm         = "Loaded form"
	LogTreeChildUIDs      = "Tree ChildUIDs called for root"
	LogTreeIsBranch       = "Tree IsBranch called"
	LogTreeCreateNode     = "Tree CreateNode called"
	LogTreeUpdateNode     = "Tree UpdateNode called"
	LogTreeUpdateNodeRoot = "Tree UpdateNode called for root"
	LogTreeSelected       = "Tree OnSelected called"
)
