package gocaptcha

type RecaptchaV2Payload struct {
	// EndpointUrl is the endpoint that has Recaptcha Protection
	EndpointUrl string

	// EndpointKey is the Recaptcha Key
	// Can be found on the Endpoint URL page
	EndpointKey string

	// IsInvisibleCaptcha Enable if endpoint has invisible Recaptcha V2
	IsInvisibleCaptcha bool

	// PageAction some site in anchor endpoint have sa param ,it's action value
	PageAction string

	// ApiDomain Domain address from which to load reCAPTCHA Enterprise.
	//
	// For example:
	//
	//• http://www.google.com/
	//
	//• http://www.recaptcha.net/
	//
	//Don't use a parameter if you don't know why it's needed.
	ApiDomain string

	// UserAgent Browser's User-Agent which is used in emulation. It is required that you use a signature of a modern browser, otherwise Google will ask you to "update your browser".
	UserAgent string
}

type RecaptchaV3Payload struct {
	// EndpointUrl is the endpoint that has Recaptcha Protection
	EndpointUrl string

	// EndpointKey is the Recaptcha Key
	// Can be found on the Endpoint URL page
	EndpointKey string

	// Action is the action name of the recaptcha, you can find it in source code of site
	Action string

	// IsEnterprise should be set if V3 Enterprise is used
	IsEnterprise bool

	// MinScore defaults to 0.3, accepted values are 0.3, 0.6, 0.9
	MinScore float32
}

type TurnstilePayload struct {
	// EndpointUrl is the endpoint that has FunCaptcha Protection
	EndpointUrl string

	// EndpointKey is the Recaptcha Key
	// Can be found on the Endpoint URL page
	EndpointKey string
}

type ImageCaptchaPayload struct {
	// Base64String is the base64 representation of the image
	Base64String string

	// CaseSensitive should be set to true if captcha is case-sensitive
	CaseSensitive bool

	// InstructionsForSolver should be set if the human solver needs additional information
	// about how to solve the captcha
	InstructionsForSolver string

	// Score 0.8 ~ 1, Identify the matching degree. If the recognition rate is not within the range, no deduction
	Score float32

	// Module Specifies the module. optional. you can follow the service documentation for supported modules
	Module string
}

type HCaptchaPayload struct {
	// EndpointUrl is the endpoint that has Recaptcha Protection
	EndpointUrl string

	// EndpointKey is the HCaptcha Key
	// Can be found on the Endpoint URL page
	EndpointKey string
}

type AntiCloudflarePayload struct {
	// EndpointUrl is the endpoint that has Recaptcha Protection
	EndpointUrl string

	// Proxy is the Proxy to get the protection solved with
	Proxy string
}
