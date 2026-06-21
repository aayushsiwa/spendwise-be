package handlers

func (h *Handler) GenerateCustomID(date string) (string, error) {
	return genIDFunc(date)
}
