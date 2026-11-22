// Команда измеряет расчёт и повтор сохранённого запроса через один канал gRPC.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
	multigroup "github.com/Hecatoncheir/pob2_engine/testdata/reference/0.0.5"
)

type measurement struct {
	FirstNS    int64   `json:"first_ns"`
	RepeatedNS []int64 `json:"repeated_ns"`
	ReplayNS   []int64 `json:"replay_ns"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var ready grpcapi.Ready
	if err := json.NewDecoder(os.Stdin).Decode(&ready); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	client, err := grpcapi.Dial(ctx, ready.Address, ready.Token)
	if err != nil {
		return err
	}
	defer client.Close()
	values := make(map[string]measurement)
	for _, name := range []string{"caster", "attacker", "minion"} {
		value, err := measure(ctx, client, name)
		if err != nil {
			return fmt.Errorf("сборка %s: %w", name, err)
		}
		values[name] = value
	}
	return json.NewEncoder(os.Stdout).Encode(values)
}

func measure(ctx context.Context, client *grpcapi.Client, name string) (measurement, error) {
	xml, err := multigroup.Builds.ReadFile(name + ".xml")
	if err != nil {
		return measurement{}, err
	}
	if _, err := client.Create(ctx, application.CreateSession{SessionID: name, BuildName: name, RequestID: "create-" + name}); err != nil {
		return measurement{}, err
	}
	opened, err := client.Apply(ctx, application.ApplyBuild{SessionID: name, ExpectedRevision: 1, RequestID: "import", XML: string(xml)})
	if err != nil {
		return measurement{}, err
	}
	command := application.CalculateBuild{SessionID: name, ExpectedRevision: opened.Revision}
	return measureCalculations(ctx, client, command)
}

func measureCalculations(ctx context.Context, client *grpcapi.Client, command application.CalculateBuild) (measurement, error) {
	const repeats = 5
	values := make([]int64, repeats+1)
	for i := range values {
		command.RequestID = fmt.Sprintf("measure-%d", i)
		elapsed, err := timedCalculation(ctx, client, command)
		if err != nil {
			return measurement{}, err
		}
		values[i] = elapsed
		command.ExpectedRevision++
	}
	command.ExpectedRevision--
	replays := make([]int64, repeats)
	for i := range replays {
		elapsed, err := timedCalculation(ctx, client, command)
		if err != nil {
			return measurement{}, err
		}
		replays[i] = elapsed
	}
	return measurement{FirstNS: values[0], RepeatedNS: values[1:], ReplayNS: replays}, nil
}

func timedCalculation(ctx context.Context, client *grpcapi.Client, command application.CalculateBuild) (int64, error) {
	started := time.Now()
	value, err := client.Calculate(ctx, command)
	elapsed := time.Since(started).Nanoseconds()
	if err != nil {
		return 0, err
	}
	result, err := application.DecodeCalculation(value)
	if err != nil || result == nil || result.Stats["FullDPS"] <= 0 || value.Revision != command.ExpectedRevision+1 {
		return 0, fmt.Errorf("не получен расчёт ожидаемой ревизии: %v", err)
	}
	return elapsed, nil
}
