package api

import (
	"context"
	"log"
	"unicode/utf8"

	"github.com/javiermm8/open-gym-vault/internal/api/gen"
)

func (s *Server) GetPersonalRecord(ctx context.Context, r gen.GetPersonalRecordRequestObject) (gen.GetPersonalRecordResponseObject, error) {
	if r.Id == "" {
		return gen.GetPersonalRecord400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{
				Error: "id is required",
			},
		}, nil
	}
	if utf8.RuneCountInString(r.Id) != 36 {
		return gen.GetPersonalRecord400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{
				Error: "id must be a valid exercise id",
			},
		}, nil
	}

	pr, err := s.store.CalculatePR(ctx, AuthenticateUserID(ctx), r.Id)
	if err != nil {
		log.Printf("500 at GetPersonalRecord: %v", err)
		return nil, err
	}
	return gen.GetPersonalRecord200JSONResponse{
		MaxWeight: &pr,
	}, nil
}
