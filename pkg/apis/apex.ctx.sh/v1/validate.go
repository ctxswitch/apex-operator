package v1

type Validator struct{}

func (v *Validator) ValidateCreate(s *Scraper) error {
	return nil
}

func (v *Validator) ValidateUpdate(old *Scraper, new *Scraper) error {
	return nil
}

func (v *Validator) ValidateDelete(s *Scraper) error {
	return nil
}
