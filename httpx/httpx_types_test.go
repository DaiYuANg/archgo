package httpx

type pingOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

type echoInput struct {
	Body struct {
		Name string `json:"name"`
	}
}

type echoOutput struct {
	Body struct {
		Name string `json:"name"`
	}
}

type customBindInput struct {
	ID    int    `query:"user_id"`
	Token string `header:"X-Token"`
}

type customBindOutput struct {
	Body struct {
		ID    int    `json:"id"`
		Token string `json:"token"`
	}
}

type paramsInput struct {
	ID    int    `query:"id"`
	Flag  bool   `query:"flag"`
	Trace string `header:"X-Trace-ID"`
}

type paramsOutput struct {
	Body struct {
		ID    int    `json:"id"`
		Flag  bool   `json:"flag"`
		Trace string `json:"trace"`
	}
}

type validatedBodyInput struct {
	Body struct {
		Name string `json:"name" validate:"required,min=3"`
	}
}

type validatedBodyOutput struct {
	Body struct {
		Name string `json:"name"`
	}
}

type validatedQueryInput struct {
	ID int `query:"id" validate:"required,min=1"`
}

type customValidatedInput struct {
	Body struct {
		Name string `json:"name" validate:"arc"`
	}
}

type humaPingOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}
