package monitor

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/google/uuid"
	"github.com/koblas/besops/lib/errs"
	"github.com/koblas/besops/models"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

type repo struct {
	db bob.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repo{db: bob.NewDB(db)}
}

func (r *repo) FindByID(ctx context.Context, id string) (*Monitor, error) {
	m, err := models.FindMonitor(ctx, r.db, id)
	if err != nil {
		return nil, errs.WrapNotFound(err, "finding monitor") //nolint:wrapcheck // WrapNotFound handles wrapping
	}
	return monitorFromModel(m), nil
}

func (r *repo) FindByUserID(ctx context.Context, userID string) ([]*Monitor, error) {
	ms, err := models.Monitors.Query(
		sm.Where(models.Monitors.Columns.UserID.EQ(sqlite.Arg(userID))),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("querying monitors by user: %w", err)
	}
	return monitorsFromModels(ms), nil
}

func (r *repo) FindByPushToken(ctx context.Context, token string) (*Monitor, error) {
	m, err := models.Monitors.Query(
		sm.Where(models.Monitors.Columns.PushToken.EQ(sqlite.Arg(token))),
	).One(ctx, r.db)
	if err != nil {
		return nil, errs.WrapNotFound(err, "finding monitor by push token") //nolint:wrapcheck // WrapNotFound handles wrapping
	}
	return monitorFromModel(m), nil
}

func (r *repo) FindAllActiveIDs(ctx context.Context) ([]string, error) {
	ms, err := models.Monitors.Query(
		sm.Where(models.Monitors.Columns.Active.EQ(sqlite.Arg(true))),
		sm.Columns(models.Monitors.Columns.Only("id")),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("querying active monitor IDs: %w", err)
	}

	ids := make([]string, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids, nil
}

func (r *repo) Create(ctx context.Context, m *Monitor) (string, error) {
	m.ID = uuid.New().String()

	_, err := models.Monitors.Insert(monitorToSetter(m)).One(ctx, r.db)
	if err != nil {
		return "", fmt.Errorf("inserting monitor: %w", err)
	}
	return m.ID, nil
}

func (r *repo) Update(ctx context.Context, m *Monitor) error {
	existing, err := models.FindMonitor(ctx, r.db, m.ID)
	if err != nil {
		return fmt.Errorf("finding monitor for update: %w", err)
	}

	setter := monitorToSetter(m)
	setter.ID = omit.Val[string]{}

	if err := existing.Update(ctx, r.db, setter); err != nil {
		return fmt.Errorf("updating monitor: %w", err)
	}
	return nil
}

func (r *repo) Delete(ctx context.Context, id string) error {
	m, err := models.FindMonitor(ctx, r.db, id)
	if err != nil {
		return fmt.Errorf("finding monitor for delete: %w", err)
	}

	if err := m.Delete(ctx, r.db); err != nil {
		return fmt.Errorf("deleting monitor: %w", err)
	}
	return nil
}

func (r *repo) GetChildren(ctx context.Context, parentID string) ([]*Monitor, error) {
	ms, err := models.Monitors.Query(
		sm.Where(models.Monitors.Columns.ParentID.EQ(sqlite.Arg(parentID))),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("querying child monitors: %w", err)
	}
	return monitorsFromModels(ms), nil
}

func monitorToSetter(m *Monitor) *models.MonitorSetter {
	s := &models.MonitorSetter{
		ID:                      omit.From(m.ID),
		Name:                    omit.From(m.Name),
		Active:                  omit.From(m.Active),
		UserID:                  omit.From(m.UserID),
		Interval:                omit.From(int64(m.Interval)),
		URL:                     omitnull.From(m.URL),
		Type:                    omit.From(m.Type),
		Weight:                  omitnull.From(int64(m.Weight)),
		Hostname:                omitnull.From(m.Hostname),
		Keyword:                 omitnull.From(m.Keyword),
		Maxretries:              omit.From(int64(m.MaxRetries)),
		IgnoreTLS:               omit.From(m.IgnoreTLS),
		UpsideDown:              omit.From(m.UpsideDown),
		Maxredirects:            omit.From(int64(m.MaxRedirects)),
		AcceptedStatuscodesJSON: omitnull.From(m.AcceptedStatusCodes),
		DNSResolveType:          omitnull.From(m.DNSResolveType),
		DNSResolveServer:        omitnull.From(m.DNSResolveServer),
		RetryInterval:           omit.From(int64(m.RetryInterval)),
		PushToken:               omitnull.From(m.PushToken),
		Method:                  omit.From(m.Method),
		Body:                    omitnull.From(m.Body),
		Headers:                 omitnull.From(m.Headers),
		BasicAuthUser:           omitnull.From(m.BasicAuthUser),
		BasicAuthPass:           omitnull.From(m.BasicAuthPass),
		Description:             omitnull.From(m.Description),
		TLSCert:                 omitnull.From(m.TLSCert),
		TLSKey:                  omitnull.From(m.TLSKey),
		TLSCa:                   omitnull.From(m.TLSCA),
		MQTTTopic:               omitnull.From(m.MQTTTopic),
		MQTTSuccessMessage:      omitnull.From(m.MQTTSuccessMessage),
		MQTTUsername:            omitnull.From(m.MQTTUsername),
		MQTTPassword:            omitnull.From(m.MQTTPassword),
		DatabaseQuery:           omitnull.From(m.DatabaseQuery),
		AuthMethod:              omitnull.From(m.AuthMethod),
		AuthWorkstation:         omitnull.From(m.AuthWorkstation),
		AuthDomain:              omitnull.From(m.AuthDomain),
		GRPCURL:                 omitnull.From(m.GRPCUrl),
		GRPCProtobuf:            omitnull.From(m.GRPCProtobuf),
		GRPCServiceName:         omitnull.From(m.GRPCServiceName),
		GRPCMethod:              omitnull.From(m.GRPCMethod),
		GRPCBody:                omitnull.From(m.GRPCBody),
		GRPCMetadata:            omitnull.From(m.GRPCMetadata),
		GRPCEnableTLS:           omit.From(m.GRPCEnableTLS),
		RadiusUsername:          omitnull.From(m.RadiusUsername),
		RadiusPassword:          omitnull.From(m.RadiusPassword),
		RadiusSecret:            omitnull.From(m.RadiusSecret),
		RadiusCalledStationID:   omitnull.From(m.RadiusCalledStation),
		RadiusCallingStationID:  omitnull.From(m.RadiusCalling),
		Game:                    omitnull.From(m.GameName),
		HTTPBodyEncoding:        omitnull.From(m.HTTPBodyEncoding),
		Timeout:                 omit.From(int64(m.Timeout)),
		InvertKeyword:           omit.From(m.InvertKeyword),
		JSONPath:                omitnull.From(m.JsonPath),
		ExpectedValue:           omitnull.From(m.ExpectedValue),
		PacketSize:              omit.From(int64(m.PacketSize)),
		ResendInterval:          omit.From(int64(m.ResendInterval)),
		KafkaProducerTopic:      omitnull.From(m.KafkaProducerTopic),
		KafkaProducerBrokers:    omitnull.From(m.KafkaProducerBrokers),
		KafkaProducerSSL:        omit.From(m.KafkaProducerSSL),
		KafkaProducerMessage:    omitnull.From(m.KafkaProducerMessage),
		RemoteBrowser:           omitnull.FromPtr(m.RemoteBrowser),
	}

	if m.Port != nil {
		port := int64(*m.Port)
		s.Port = omitnull.FromPtr(&port)
	} else {
		s.Port = omitnull.FromPtr[int64](nil)
	}
	s.ProxyID = omitnull.FromPtr(m.ProxyID)
	s.ParentID = omitnull.FromPtr(m.ParentID)
	s.ExpiryNotification = omitnull.From(m.ExpiryNotification)

	return s
}

func monitorFromModel(m *models.Monitor) *Monitor {
	mon := &Monitor{
		ID:                   m.ID,
		Name:                 m.Name,
		Active:               m.Active,
		UserID:               m.UserID,
		Interval:             int(m.Interval),
		URL:                  m.URL.GetOrZero(),
		Type:                 m.Type,
		Weight:               int(m.Weight.GetOrZero()),
		Hostname:             m.Hostname.GetOrZero(),
		CreatedDate:          m.CreatedDate,
		Keyword:              m.Keyword.GetOrZero(),
		MaxRetries:           int(m.Maxretries),
		IgnoreTLS:            m.IgnoreTLS,
		UpsideDown:           m.UpsideDown,
		MaxRedirects:         int(m.Maxredirects),
		AcceptedStatusCodes:  m.AcceptedStatuscodesJSON.GetOrZero(),
		DNSResolveType:       m.DNSResolveType.GetOrZero(),
		DNSResolveServer:     m.DNSResolveServer.GetOrZero(),
		DNSLastResult:        m.DNSLastResult.GetOrZero(),
		RetryInterval:        int(m.RetryInterval),
		PushToken:            m.PushToken.GetOrZero(),
		Method:               m.Method,
		Body:                 m.Body.GetOrZero(),
		Headers:              m.Headers.GetOrZero(),
		BasicAuthUser:        m.BasicAuthUser.GetOrZero(),
		BasicAuthPass:        m.BasicAuthPass.GetOrZero(),
		Description:          m.Description.GetOrZero(),
		TLSCert:              m.TLSCert.GetOrZero(),
		TLSKey:               m.TLSKey.GetOrZero(),
		TLSCA:                m.TLSCa.GetOrZero(),
		MQTTTopic:            m.MQTTTopic.GetOrZero(),
		MQTTSuccessMessage:   m.MQTTSuccessMessage.GetOrZero(),
		MQTTUsername:         m.MQTTUsername.GetOrZero(),
		MQTTPassword:         m.MQTTPassword.GetOrZero(),
		DatabaseQuery:        m.DatabaseQuery.GetOrZero(),
		AuthMethod:           m.AuthMethod.GetOrZero(),
		AuthWorkstation:      m.AuthWorkstation.GetOrZero(),
		AuthDomain:           m.AuthDomain.GetOrZero(),
		GRPCUrl:              m.GRPCURL.GetOrZero(),
		GRPCProtobuf:         m.GRPCProtobuf.GetOrZero(),
		GRPCServiceName:      m.GRPCServiceName.GetOrZero(),
		GRPCMethod:           m.GRPCMethod.GetOrZero(),
		GRPCBody:             m.GRPCBody.GetOrZero(),
		GRPCMetadata:         m.GRPCMetadata.GetOrZero(),
		GRPCEnableTLS:        m.GRPCEnableTLS,
		RadiusUsername:       m.RadiusUsername.GetOrZero(),
		RadiusPassword:       m.RadiusPassword.GetOrZero(),
		RadiusSecret:         m.RadiusSecret.GetOrZero(),
		RadiusCalledStation:  m.RadiusCalledStationID.GetOrZero(),
		RadiusCalling:        m.RadiusCallingStationID.GetOrZero(),
		GameName:             m.Game.GetOrZero(),
		HTTPBodyEncoding:     m.HTTPBodyEncoding.GetOrZero(),
		Timeout:              int(m.Timeout),
		InvertKeyword:        m.InvertKeyword,
		JsonPath:             m.JSONPath.GetOrZero(),
		ExpectedValue:        m.ExpectedValue.GetOrZero(),
		PacketSize:           int(m.PacketSize),
		ResendInterval:       int(m.ResendInterval),
		KafkaProducerTopic:   m.KafkaProducerTopic.GetOrZero(),
		KafkaProducerBrokers: m.KafkaProducerBrokers.GetOrZero(),
		KafkaProducerSSL:     m.KafkaProducerSSL,
		KafkaProducerMessage: m.KafkaProducerMessage.GetOrZero(),
	}

	if v, ok := m.Port.Get(); ok {
		port := int(v)
		mon.Port = &port
	}
	if v, ok := m.ProxyID.Get(); ok {
		mon.ProxyID = &v
	}
	if v, ok := m.ParentID.Get(); ok {
		mon.ParentID = &v
	}
	if v, ok := m.ExpiryNotification.Get(); ok {
		mon.ExpiryNotification = v
	}
	if v, ok := m.RemoteBrowser.Get(); ok {
		mon.RemoteBrowser = &v
	}

	return mon
}

func monitorsFromModels(ms models.MonitorSlice) []*Monitor {
	result := make([]*Monitor, len(ms))
	for i, m := range ms {
		result[i] = monitorFromModel(m)
	}
	return result
}
