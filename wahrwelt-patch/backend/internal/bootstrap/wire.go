//go:build wireinject

package bootstrap

import (
	"context"

	"github.com/TakuyaYagam1/wahrwelt/config"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/middleware"
	httpv1 "github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/v1"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/secondary/filesystem"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/secondary/password"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/secondary/postgres"
	redisadapter "github.com/TakuyaYagam1/wahrwelt/internal/adapter/secondary/redis"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/secondary/registry"
	"github.com/TakuyaYagam1/wahrwelt/internal/port/outbound"
	"github.com/TakuyaYagam1/wahrwelt/internal/usecase"
	"github.com/goforj/wire"
)

var appSet = wire.NewSet(
	postgres.NewPool,
	postgres.NewInvoiceSearchPool,
	postgres.NewStore,
	redisadapter.New,
	registry.NewClient,
	password.New,
	provideArchiveFiles,
	providePreviewFiles,
	usecase.NewCase,
	usecase.NewManifest,
	usecase.NewCertificate,
	usecase.NewInvoice,
	usecase.NewInspection,
	usecase.NewContract,
	usecase.NewQuota,
	usecase.NewWaybill,
	usecase.NewPreview,
	usecase.NewGatePass,
	provideLicenses,
	provideAuth,
	provideRouteSlots,
	provideHandler,
	middleware.NewSecurityHandler,
	httpv1.NewServer,
	NewApp,
	wire.Bind(new(outbound.UserRepository), new(*postgres.Store)),
	wire.Bind(new(outbound.CaseRepository), new(*postgres.Store)),
	wire.Bind(new(outbound.RouteSlotRepository), new(*postgres.Store)),
	wire.Bind(new(outbound.ManifestRepository), new(*postgres.Store)),
	wire.Bind(new(outbound.CertificateRepository), new(*postgres.Store)),
	wire.Bind(new(outbound.InvoiceRepository), new(*postgres.Store)),
	wire.Bind(new(outbound.InspectionRepository), new(*postgres.Store)),
	wire.Bind(new(outbound.LicenseRepository), new(*postgres.Store)),
	wire.Bind(new(outbound.ContractRepository), new(*postgres.Store)),
	wire.Bind(new(outbound.QuotaRepository), new(*postgres.Store)),
	wire.Bind(new(outbound.WaybillRepository), new(*postgres.Store)),
	wire.Bind(new(outbound.PreviewRepository), new(*postgres.Store)),
	wire.Bind(new(outbound.GatePassRepository), new(*postgres.Store)),
	wire.Bind(new(outbound.WaybillFiles), new(*filesystem.Archive)),
	wire.Bind(new(outbound.PreviewFiles), new(*filesystem.Preview)),
	wire.Bind(new(outbound.CertificateVerifier), new(*registry.Client)),
	wire.Bind(new(outbound.SessionStore), new(*redisadapter.Store)),
	wire.Bind(new(outbound.ApprovalStore), new(*redisadapter.Store)),
	wire.Bind(new(outbound.PasswordHasher), new(*password.Hasher)),
)

func initializeApp(context.Context, config.Config) (*App, func(), error) {
	wire.Build(appSet)
	return nil, nil, nil
}

func provideAuth(cfg config.Config, users outbound.UserRepository, sessions outbound.SessionStore, hasher outbound.PasswordHasher) *usecase.Auth {
	return usecase.NewAuth(users, sessions, hasher, cfg.SessionTTL)
}

func provideRouteSlots(cfg config.Config, cases outbound.CaseRepository, users outbound.UserRepository, slots outbound.RouteSlotRepository, approvals outbound.ApprovalStore) *usecase.RouteSlot {
	return usecase.NewRouteSlot(cases, users, slots, approvals, cfg.ApprovalTTL)
}

func provideLicenses(cfg config.Config, repository outbound.LicenseRepository) *usecase.License {
	return usecase.NewLicense(repository, cfg.LicenseTokenSecret, cfg.LicenseTokenTTL)
}

func provideArchiveFiles() *filesystem.Archive {
	return filesystem.NewArchive("/var/lib/wahrwelt/archive")
}

func providePreviewFiles() *filesystem.Preview {
	return filesystem.NewPreview("/var/lib/wahrwelt/preview")
}

func provideHandler(cfg config.Config, auth *usecase.Auth, cases *usecase.Case, slots *usecase.RouteSlot, manifests *usecase.Manifest, certificates *usecase.Certificate, invoices *usecase.Invoice, inspections *usecase.Inspection, licenses *usecase.License, contracts *usecase.Contract, quota *usecase.Quota, waybill *usecase.Waybill, preview *usecase.Preview, gatePass *usecase.GatePass) *httpv1.Handler {
	return httpv1.NewHandler(auth, cases, slots, httpv1.FeatureSet{Manifest: manifests, Certificate: certificates, Invoice: invoices, Inspection: inspections, License: licenses, Contract: contracts, Quota: quota, Waybill: waybill, Preview: preview, GatePass: gatePass}, cfg.SessionTTL, cfg.CookieSecure)
}
