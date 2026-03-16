// Package main provides REST API service for go-indigo chemistry toolkit
// Mimics the EPAM Indigo Python Service API
// coding=utf-8
// @Project : go-indigo
// @Author  : chengxiang.luo
// @Email   : chengxiang.luo@foxmail.com
// @File    : model.go
package main

// --- Request Models ---

// StructRequest is the common request body for structure-related endpoints
type StructRequest struct {
	Struct      string            `json:"struct"`
	InputFormat string            `json:"input_format,omitempty"`
	Options     map[string]string `json:"options,omitempty"`
}

// ConvertRequest extends StructRequest with output format
type ConvertRequest struct {
	StructRequest
	OutputFormat string `json:"output_format"`
}

// RenderRequest extends StructRequest with render-specific options
type RenderRequest struct {
	StructRequest
	OutputFormat string `json:"output_format,omitempty"` // png, svg, pdf
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
}

// CalculateRequest extends StructRequest with property selection
type CalculateRequest struct {
	StructRequest
	Properties []string `json:"properties,omitempty"` // which properties to calculate, empty = all
}

// CheckRequest extends StructRequest with check types
type CheckRequest struct {
	StructRequest
	Types []string `json:"types,omitempty"` // valence, ambiguous_h, etc.
}

// --- Response Models ---

// StructResponse is the common response body for structure-related endpoints
type StructResponse struct {
	Struct         string `json:"struct"`
	Format         string `json:"format"`
	OriginalFormat string `json:"original_format,omitempty"`
}

// CalculateResponse contains calculated properties
type CalculateResponse struct {
	MolecularWeight   *float64 `json:"molecular-weight,omitempty"`
	MostAbundantMass  *float64 `json:"most-abundant-mass,omitempty"`
	MonoisotopicMass  *float64 `json:"monoisotopic-mass,omitempty"`
	MassComposition   string   `json:"mass-composition,omitempty"`
	GrossFormula      string   `json:"gross-formula,omitempty"`
	MolecularFormula  string   `json:"molecular-formula,omitempty"`
	TPSA              *float64 `json:"tpsa,omitempty"`
	NumRotatableBonds *int     `json:"num-rotatable-bonds,omitempty"`
	NumAtoms          *int     `json:"num-atoms,omitempty"`
	NumBonds          *int     `json:"num-bonds,omitempty"`
	NumHeavyAtoms     *int     `json:"num-heavy-atoms,omitempty"`
	NumComponents     *int     `json:"num-components,omitempty"`
	NumSSSR           *int     `json:"num-sssr,omitempty"`
}

// RenderResponse contains rendered image data
type RenderResponse struct {
	Image  string `json:"image"`  // base64 encoded image
	Format string `json:"format"` // png, svg, pdf
}

// CheckResponse contains validation results
type CheckResponse struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
	Struct string   `json:"struct,omitempty"`
	Format string   `json:"format,omitempty"`
}

// InfoResponse contains library version info
type InfoResponse struct {
	IndigoVersion  string `json:"indigo_version"`
	ServiceName    string `json:"service_name"`
	ServiceVersion string `json:"service_version"`
}

// ErrorResponse contains error information
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
