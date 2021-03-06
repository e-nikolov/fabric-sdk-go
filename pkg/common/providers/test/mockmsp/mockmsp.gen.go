// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp (interfaces: IdentityConfig,IdentityManager,Providers)

package mockmsp

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	msp "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
)

// MockIdentityConfig is a mock of IdentityConfig interface
type MockIdentityConfig struct {
	ctrl     *gomock.Controller
	recorder *MockIdentityConfigMockRecorder
}

// MockIdentityConfigMockRecorder is the mock recorder for MockIdentityConfig
type MockIdentityConfigMockRecorder struct {
	mock *MockIdentityConfig
}

// NewMockIdentityConfig creates a new mock instance
func NewMockIdentityConfig(ctrl *gomock.Controller) *MockIdentityConfig {
	mock := &MockIdentityConfig{ctrl: ctrl}
	mock.recorder = &MockIdentityConfigMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *MockIdentityConfig) EXPECT() *MockIdentityConfigMockRecorder {
	return _m.recorder
}

// CAClientCert mocks base method
func (_m *MockIdentityConfig) CAClientCert(_param0 string) ([]byte, error) {
	ret := _m.ctrl.Call(_m, "CAClientCert", _param0)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CAClientCert indicates an expected call of CAClientCert
func (_mr *MockIdentityConfigMockRecorder) CAClientCert(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "CAClientCert", reflect.TypeOf((*MockIdentityConfig)(nil).CAClientCert), arg0)
}

// CAClientKey mocks base method
func (_m *MockIdentityConfig) CAClientKey(_param0 string) ([]byte, error) {
	ret := _m.ctrl.Call(_m, "CAClientKey", _param0)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CAClientKey indicates an expected call of CAClientKey
func (_mr *MockIdentityConfigMockRecorder) CAClientKey(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "CAClientKey", reflect.TypeOf((*MockIdentityConfig)(nil).CAClientKey), arg0)
}

// CAConfig mocks base method
func (_m *MockIdentityConfig) CAConfig(_param0 string) (*msp.CAConfig, error) {
	ret := _m.ctrl.Call(_m, "CAConfig", _param0)
	ret0, _ := ret[0].(*msp.CAConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CAConfig indicates an expected call of CAConfig
func (_mr *MockIdentityConfigMockRecorder) CAConfig(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "CAConfig", reflect.TypeOf((*MockIdentityConfig)(nil).CAConfig), arg0)
}

// CAKeyStorePath mocks base method
func (_m *MockIdentityConfig) CAKeyStorePath() string {
	ret := _m.ctrl.Call(_m, "CAKeyStorePath")
	ret0, _ := ret[0].(string)
	return ret0
}

// CAKeyStorePath indicates an expected call of CAKeyStorePath
func (_mr *MockIdentityConfigMockRecorder) CAKeyStorePath() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "CAKeyStorePath", reflect.TypeOf((*MockIdentityConfig)(nil).CAKeyStorePath))
}

// CAServerCerts mocks base method
func (_m *MockIdentityConfig) CAServerCerts(_param0 string) ([][]byte, error) {
	ret := _m.ctrl.Call(_m, "CAServerCerts", _param0)
	ret0, _ := ret[0].([][]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CAServerCerts indicates an expected call of CAServerCerts
func (_mr *MockIdentityConfigMockRecorder) CAServerCerts(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "CAServerCerts", reflect.TypeOf((*MockIdentityConfig)(nil).CAServerCerts), arg0)
}

// Client mocks base method
func (_m *MockIdentityConfig) Client() (*msp.ClientConfig, error) {
	ret := _m.ctrl.Call(_m, "Client")
	ret0, _ := ret[0].(*msp.ClientConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Client indicates an expected call of Client
func (_mr *MockIdentityConfigMockRecorder) Client() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Client", reflect.TypeOf((*MockIdentityConfig)(nil).Client))
}

// CredentialStorePath mocks base method
func (_m *MockIdentityConfig) CredentialStorePath() string {
	ret := _m.ctrl.Call(_m, "CredentialStorePath")
	ret0, _ := ret[0].(string)
	return ret0
}

// CredentialStorePath indicates an expected call of CredentialStorePath
func (_mr *MockIdentityConfigMockRecorder) CredentialStorePath() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "CredentialStorePath", reflect.TypeOf((*MockIdentityConfig)(nil).CredentialStorePath))
}

// MockIdentityManager is a mock of IdentityManager interface
type MockIdentityManager struct {
	ctrl     *gomock.Controller
	recorder *MockIdentityManagerMockRecorder
}

// MockIdentityManagerMockRecorder is the mock recorder for MockIdentityManager
type MockIdentityManagerMockRecorder struct {
	mock *MockIdentityManager
}

// NewMockIdentityManager creates a new mock instance
func NewMockIdentityManager(ctrl *gomock.Controller) *MockIdentityManager {
	mock := &MockIdentityManager{ctrl: ctrl}
	mock.recorder = &MockIdentityManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *MockIdentityManager) EXPECT() *MockIdentityManagerMockRecorder {
	return _m.recorder
}

// GetSigningIdentity mocks base method
func (_m *MockIdentityManager) GetSigningIdentity(_param0 string) (msp.SigningIdentity, error) {
	ret := _m.ctrl.Call(_m, "GetSigningIdentity", _param0)
	ret0, _ := ret[0].(msp.SigningIdentity)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSigningIdentity indicates an expected call of GetSigningIdentity
func (_mr *MockIdentityManagerMockRecorder) GetSigningIdentity(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "GetSigningIdentity", reflect.TypeOf((*MockIdentityManager)(nil).GetSigningIdentity), arg0)
}

// MockProviders is a mock of Providers interface
type MockProviders struct {
	ctrl     *gomock.Controller
	recorder *MockProvidersMockRecorder
}

// MockProvidersMockRecorder is the mock recorder for MockProviders
type MockProvidersMockRecorder struct {
	mock *MockProviders
}

// NewMockProviders creates a new mock instance
func NewMockProviders(ctrl *gomock.Controller) *MockProviders {
	mock := &MockProviders{ctrl: ctrl}
	mock.recorder = &MockProvidersMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *MockProviders) EXPECT() *MockProvidersMockRecorder {
	return _m.recorder
}

// IdentityConfig mocks base method
func (_m *MockProviders) IdentityConfig() msp.IdentityConfig {
	ret := _m.ctrl.Call(_m, "IdentityConfig")
	ret0, _ := ret[0].(msp.IdentityConfig)
	return ret0
}

// IdentityConfig indicates an expected call of IdentityConfig
func (_mr *MockProvidersMockRecorder) IdentityConfig() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "IdentityConfig", reflect.TypeOf((*MockProviders)(nil).IdentityConfig))
}

// IdentityManager mocks base method
func (_m *MockProviders) IdentityManager(_param0 string) (msp.IdentityManager, bool) {
	ret := _m.ctrl.Call(_m, "IdentityManager", _param0)
	ret0, _ := ret[0].(msp.IdentityManager)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// IdentityManager indicates an expected call of IdentityManager
func (_mr *MockProvidersMockRecorder) IdentityManager(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "IdentityManager", reflect.TypeOf((*MockProviders)(nil).IdentityManager), arg0)
}

// UserStore mocks base method
func (_m *MockProviders) UserStore() msp.UserStore {
	ret := _m.ctrl.Call(_m, "UserStore")
	ret0, _ := ret[0].(msp.UserStore)
	return ret0
}

// UserStore indicates an expected call of UserStore
func (_mr *MockProvidersMockRecorder) UserStore() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "UserStore", reflect.TypeOf((*MockProviders)(nil).UserStore))
}
