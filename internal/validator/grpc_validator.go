package validator

import (
	"fmt"
	"sync"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
)

type GRPCValidator struct {
	logger      logging.Logger
	fileDescs   map[string]*desc.FileDescriptor
	msgDescs    map[string]*desc.MessageDescriptor
	methodDescs map[string]*desc.MethodDescriptor
	mu          sync.RWMutex
}

type GRPCValidationConfig struct {
	ProtoFiles       []string          `yaml:"proto_files" json:"proto_files"`
	ProtoImportPaths []string          `yaml:"proto_import_paths" json:"proto_import_paths"`
	ServiceMappings  map[string]string `yaml:"service_mappings" json:"service_mappings"`
}

type GRPCValidationResult struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Message string   `json:"message,omitempty"`
}

func NewGRPCValidator(config *GRPCValidationConfig, logger logging.Logger) (*GRPCValidator, error) {
	validator := &GRPCValidator{
		logger:      logger,
		fileDescs:   make(map[string]*desc.FileDescriptor),
		msgDescs:    make(map[string]*desc.MessageDescriptor),
		methodDescs: make(map[string]*desc.MethodDescriptor),
	}

	if config != nil && len(config.ProtoFiles) > 0 {
		err := validator.LoadProtoFiles(config.ProtoFiles, config.ProtoImportPaths)
		if err != nil {
			return nil, err
		}
	}

	return validator, nil
}

func (v *GRPCValidator) LoadProtoFiles(protoFiles []string, importPaths []string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	parser := protoparse.Parser{
		ImportPaths:           importPaths,
		IncludeSourceCodeInfo: true,
	}

	// Parse proto files
	fileDescs, err := parser.ParseFiles(protoFiles...)
	if err != nil {
		return fmt.Errorf("error parsing proto files: %w", err)
	}

	// Index message and method descriptors
	for _, fileDesc := range fileDescs {
		v.fileDescs[fileDesc.GetName()] = fileDesc

		for _, msgDesc := range fileDesc.GetMessageTypes() {
			v.indexMessageType(msgDesc)
		}

		for _, svcDesc := range fileDesc.GetServices() {
			for _, methodDesc := range svcDesc.GetMethods() {
				fullMethodName := svcDesc.GetFullyQualifiedName() + "." + methodDesc.GetName()
				v.methodDescs[fullMethodName] = methodDesc

				// Also store with simplified name for easier lookup
				simpleName := svcDesc.GetName() + "." + methodDesc.GetName()
				v.methodDescs[simpleName] = methodDesc
			}
		}
	}

	v.logger.Info("Loaded gRPC proto definitions",
		"files", len(fileDescs),
		"messages", len(v.msgDescs),
		"methods", len(v.methodDescs))

	return nil
}

func (v *GRPCValidator) indexMessageType(msgDesc *desc.MessageDescriptor) {
	fullName := msgDesc.GetFullyQualifiedName()
	v.msgDescs[fullName] = msgDesc

	// Also add with simple name for easier lookups
	simpleName := msgDesc.GetName()
	v.msgDescs[simpleName] = msgDesc

	// Index nested message types
	for _, nestedMsg := range msgDesc.GetNestedMessageTypes() {
		v.indexMessageType(nestedMsg)
	}
}

func (v *GRPCValidator) ValidateMessage(msgTypeName string, data []byte) (*GRPCValidationResult, error) {
	v.mu.RLock()
	msgDesc, exists := v.msgDescs[msgTypeName]
	v.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("message type %s not found", msgTypeName)
	}

	result := &GRPCValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	// Create a dynamic message from the descriptor
	msg := dynamic.NewMessage(msgDesc)

	// Unmarshal JSON data into the message
	if err := msg.UnmarshalJSON(data); err != nil {
		result.Valid = false
		result.Message = "Invalid message format"
		result.Errors = append(result.Errors, err.Error())
		return result, nil
	}

	// Validate the message against its schema
	if err := msg.ValidateAll(); err != nil {
		result.Valid = false
		result.Message = "Message validation failed"
		result.Errors = append(result.Errors, err.Error())
	}

	return result, nil
}

func (v *GRPCValidator) ValidateMethodRequest(methodName string, data []byte) (*GRPCValidationResult, error) {
	v.mu.RLock()
	methodDesc, exists := v.methodDescs[methodName]
	v.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("method %s not found", methodName)
	}

	inputMsgDesc := methodDesc.GetInputType()
	if inputMsgDesc == nil {
		return nil, fmt.Errorf("input message descriptor not found for method %s", methodName)
	}

	return v.ValidateMessage(inputMsgDesc.GetFullyQualifiedName(), data)
}

func (v *GRPCValidator) ValidateMethodResponse(methodName string, data []byte) (*GRPCValidationResult, error) {
	v.mu.RLock()
	methodDesc, exists := v.methodDescs[methodName]
	v.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("method %s not found", methodName)
	}

	outputMsgDesc := methodDesc.GetOutputType()
	if outputMsgDesc == nil {
		return nil, fmt.Errorf("output message descriptor not found for method %s", methodName)
	}

	return v.ValidateMessage(outputMsgDesc.GetFullyQualifiedName(), data)
}

func (v *GRPCValidator) GetMessageTypes() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	msgTypes := make([]string, 0, len(v.msgDescs))
	for msgType := range v.msgDescs {
		msgTypes = append(msgTypes, msgType)
	}

	return msgTypes
}

func (v *GRPCValidator) GetMethodNames() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	methodNames := make([]string, 0, len(v.methodDescs))
	for methodName := range v.methodDescs {
		methodNames = append(methodNames, methodName)
	}

	return methodNames
}
