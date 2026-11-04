package usecase

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"
	"text/template/parse"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/TakuyaYagam1/wahrwelt/internal/port/outbound"
	"uuid"
)

const (
	maxContractTemplateBytes = 512
	maxContractPreviewBytes  = 4096
)

type Contract struct {
	repository outbound.ContractRepository
}

func NewContract(repository outbound.ContractRepository) *Contract {
	return &Contract{repository: repository}
}

func (c *Contract) Create(ctx context.Context, ownerID uuid.UUID, crewName, service, crewPhrase string, startsAt, endsAt time.Time) (domain.Contract, error) {
	if ownerID == (uuid.UUID{}) {
		return domain.Contract{}, domain.ErrUnauthorized
	}
	crewName = strings.TrimSpace(crewName)
	service = strings.TrimSpace(service)
	crewPhrase = strings.TrimSpace(crewPhrase)
	if crewName == "" || len(crewName) > 120 || service == "" || len(service) > 240 || crewPhrase == "" || len(crewPhrase) > 240 || startsAt.IsZero() || endsAt.IsZero() || !startsAt.Before(endsAt) {
		return domain.Contract{}, domain.ErrInvalidInput
	}
	id := uuid.New()
	item := domain.Contract{
		ID:         id,
		OwnerID:    ownerID,
		Number:     "CTR-" + id.String(),
		CrewName:   crewName,
		Service:    service,
		CrewPhrase: crewPhrase,
		StartsAt:   startsAt.UTC(),
		EndsAt:     endsAt.UTC(),
		CreatedAt:  time.Now().UTC(),
	}
	created, err := c.repository.CreateContract(ctx, item)
	if err != nil {
		return domain.Contract{}, fmt.Errorf("create service contract: %w", err)
	}
	return created, nil
}

func (c *Contract) Get(ctx context.Context, ownerID uuid.UUID, number string) (domain.Contract, error) {
	if ownerID == (uuid.UUID{}) {
		return domain.Contract{}, domain.ErrUnauthorized
	}
	item, err := c.repository.GetContractByNumber(ctx, strings.TrimSpace(number))
	if err != nil {
		return domain.Contract{}, fmt.Errorf("get service contract: %w", err)
	}
	if item.OwnerID != ownerID {
		return domain.Contract{}, domain.ErrNotFound
	}
	return item, nil
}

func (c *Contract) PublicCard(ctx context.Context, number string) (domain.ContractCard, error) {
	item, err := c.repository.GetContractByNumber(ctx, strings.TrimSpace(number))
	if err != nil {
		return domain.ContractCard{}, fmt.Errorf("get service contract card: %w", err)
	}
	return domain.ContractCard{
		Number:   item.Number,
		CrewName: item.CrewName,
		Service:  item.Service,
		StartsAt: item.StartsAt,
		EndsAt:   item.EndsAt,
	}, nil
}

func (c *Contract) Preview(ctx context.Context, number, source string) (domain.ContractPreview, error) {
	if len(source) == 0 || len(source) > maxContractTemplateBytes {
		return domain.ContractPreview{}, domain.ErrInvalidInput
	}
	item, err := c.repository.GetContractByNumber(ctx, strings.TrimSpace(number))
	if err != nil {
		return domain.ContractPreview{}, fmt.Errorf("get service contract for preview: %w", err)
	}
	parsed, err := template.New("contract-preview").Option("missingkey=error").Parse(source)
	if err != nil || !contractTemplateAllowed(parsed.Root) {
		return domain.ContractPreview{}, domain.ErrInvalidInput
	}
	output := &boundedPreviewWriter{limit: maxContractPreviewBytes}
	if err := parsed.Execute(output, item); err != nil {
		return domain.ContractPreview{}, domain.ErrInvalidInput
	}
	return domain.ContractPreview{Rendered: output.String()}, nil
}

func contractTemplateAllowed(root *parse.ListNode) bool {
	if root == nil {
		return false
	}
	for _, node := range root.Nodes {
		switch item := node.(type) {
		case *parse.TextNode:
		case *parse.ActionNode:
			if item.Pipe == nil || len(item.Pipe.Decl) != 0 || len(item.Pipe.Cmds) != 1 || item.Pipe.IsAssign {
				return false
			}
			command := item.Pipe.Cmds[0]
			if len(command.Args) != 1 {
				return false
			}
			switch command.Args[0].(type) {
			case *parse.FieldNode, *parse.DotNode:
			default:
				return false
			}
		default:
			return false
		}
	}
	return true
}

type boundedPreviewWriter struct {
	buffer bytes.Buffer
	limit  int
}

func (w *boundedPreviewWriter) Write(value []byte) (int, error) {
	remaining := w.limit - w.buffer.Len()
	if len(value) > remaining {
		if remaining > 0 {
			_, _ = w.buffer.Write(value[:remaining])
		}
		return 0, fmt.Errorf("preview output exceeds limit")
	}
	return w.buffer.Write(value)
}

func (w *boundedPreviewWriter) String() string {
	return w.buffer.String()
}
