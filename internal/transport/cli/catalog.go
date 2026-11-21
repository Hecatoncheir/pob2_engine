package cli

import (
	"context"
	"fmt"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
)

func (r runner) catalog(ctx context.Context, args []string) error {
	fs := r.flags("catalog")
	connection := connectFlags(fs)
	query := application.CatalogQuery{}
	fs.StringVar(&query.Kind, "kind", "gems", "gems, classes, items или passives")
	fs.StringVar(&query.Query, "query", "", "поиск по ID, имени и псевдонимам")
	fs.IntVar(&query.Offset, "offset", 0, "смещение от начала найденных записей, до 10000")
	fs.IntVar(&query.Limit, "limit", 20, "число записей, до 100")
	if err := parse(fs, args); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, connection.timeout)
	defer cancel()
	client, err := connection.dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	value, err := client.Catalog(ctx, query)
	if err != nil {
		return err
	}
	return r.write(value)
}

func (r runner) previewSupport(ctx context.Context, args []string) error {
	fs := r.flags("preview-support")
	connection := connectFlags(fs)
	command := application.PreviewSupport{}
	fs.StringVar(&command.SessionID, "session", "", "явно выбранный session_id")
	fs.Uint64Var(&command.ExpectedRevision, "revision", 0, "ожидаемая ревизия")
	fs.StringVar(&command.SupportID, "support", "", "точный ID или псевдоним поддержки из каталога")
	if err := parse(fs, args); err != nil {
		return err
	}
	if command.SessionID == "" || command.SupportID == "" {
		return &UsageError{Err: fmt.Errorf("нужны --session и --support")}
	}
	ctx, cancel := context.WithTimeout(ctx, connection.timeout)
	defer cancel()
	client, err := connection.dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	value, err := client.PreviewSupport(ctx, command)
	if err != nil {
		return err
	}
	if err := r.write(value); err != nil {
		return err
	}
	return resultStatus(value.Status)
}
