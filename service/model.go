// Package main provides REST API service for go-indigo chemistry toolkit
// Mimics the EPAM Indigo Python Service API
// coding=utf-8
// @Project : go-indigo
// @Author  : chengxiang.luo
// @Email   : chengxiang.luo@foxmail.com
// @File    : model.go
package main

// --- Request Models ---

// StructRequest is the common request body for structure-related endpoints.
// @Description Common request body containing a chemical structure string and optional parameters.
type StructRequest struct {
	Struct      string            `json:"struct" example:"c1ccccc1" binding:"required"`             // Chemical structure string (SMILES, Molfile, CML, etc.)
	InputFormat string            `json:"input_format,omitempty" example:"auto"`                    // Input format hint: auto, smiles, molfile, cml, inchi
	Options     map[string]string `json:"options,omitempty" swaggertype:"object,string"`            // Additional Indigo options as key-value pairs
}

// ConvertRequest extends StructRequest with an output format.
// @Description Request body for format conversion. Requires struct and output_format.
type ConvertRequest struct {
	Struct       string            `json:"struct" example:"c1ccccc1" binding:"required"`            // Chemical structure string
	InputFormat  string            `json:"input_format,omitempty" example:"auto"`                   // Input format hint
	Options      map[string]string `json:"options,omitempty" swaggertype:"object,string"`           // Additional Indigo options
	OutputFormat string            `json:"output_format" example:"molfile" binding:"required"`      // Target format: smiles, molfile, cml, cdxml, json, ket
}

// RenderRequest extends StructRequest with render-specific options.
// @Description Request body for rendering a structure to an image.
type RenderRequest struct {
	Struct       string            `json:"struct" example:"c1ccccc1" binding:"required"`            // Chemical structure string
	InputFormat  string            `json:"input_format,omitempty" example:"auto"`                   // Input format hint
	Options      map[string]string `json:"options,omitempty" swaggertype:"object,string"`           // Additional render options (e.g. render-bond-length)
	OutputFormat string            `json:"output_format,omitempty" example:"png"`                   // Image format: png, svg, pdf (default: png)
	Width        int               `json:"width,omitempty" example:"800"`                           // Image width in pixels
	Height       int               `json:"height,omitempty" example:"600"`                          // Image height in pixels
}

// CalculateRequest extends StructRequest with property selection.
// @Description Request body for calculating molecular properties.
type CalculateRequest struct {
	Struct      string            `json:"struct" example:"CCO" binding:"required"`                  // Chemical structure string
	InputFormat string            `json:"input_format,omitempty" example:"auto"`                    // Input format hint
	Options     map[string]string `json:"options,omitempty" swaggertype:"object,string"`            // Additional Indigo options
	Properties  []string          `json:"properties,omitempty" example:"molecular-weight,gross-formula"` // Properties to calculate; empty = all
}

// CheckRequest extends StructRequest with check types.
// @Description Request body for validating a chemical structure.
type CheckRequest struct {
	Struct      string            `json:"struct" example:"CCO" binding:"required"`                  // Chemical structure string
	InputFormat string            `json:"input_format,omitempty" example:"auto"`                    // Input format hint
	Options     map[string]string `json:"options,omitempty" swaggertype:"object,string"`            // Additional Indigo options
	Types       []string          `json:"types,omitempty"`                                          // Validation types: valence, ambiguous_h, etc.
}

// --- Response Models ---

// StructResponse is the common response body for structure-related endpoints.
// @Description Response containing a chemical structure string in the requested format.
type StructResponse struct {
	Struct         string `json:"struct" example:"c1ccccc1"`           // Result structure string
	Format         string `json:"format" example:"smiles"`             // Output format
	OriginalFormat string `json:"original_format,omitempty"`           // Detected input format
}

// CalculateResponse contains calculated molecular properties.
// @Description Response containing calculated molecular properties. Null fields were not requested or could not be computed.
type CalculateResponse struct {
	MolecularWeight   *float64 `json:"molecular-weight,omitempty" example:"46.069"`    // Molecular weight (g/mol)
	MostAbundantMass  *float64 `json:"most-abundant-mass,omitempty" example:"46.042"`  // Most abundant mass
	MonoisotopicMass  *float64 `json:"monoisotopic-mass,omitempty" example:"46.042"`   // Monoisotopic mass
	MassComposition   string   `json:"mass-composition,omitempty"`                      // Elemental mass composition
	GrossFormula      string   `json:"gross-formula,omitempty" example:"C2 H6 O"`      // Gross formula
	MolecularFormula  string   `json:"molecular-formula,omitempty" example:"C2H6O"`    // Molecular formula
	TPSA              *float64 `json:"tpsa,omitempty" example:"20.23"`                  // Topological polar surface area
	NumRotatableBonds *int     `json:"num-rotatable-bonds,omitempty" example:"0"`       // Number of rotatable bonds
	NumAtoms          *int     `json:"num-atoms,omitempty" example:"9"`                 // Total atom count
	NumBonds          *int     `json:"num-bonds,omitempty" example:"8"`                 // Total bond count
	NumHeavyAtoms     *int     `json:"num-heavy-atoms,omitempty" example:"3"`           // Heavy (non-hydrogen) atom count
	NumComponents     *int     `json:"num-components,omitempty" example:"1"`            // Number of connected components
	NumSSSR           *int     `json:"num-sssr,omitempty" example:"0"`                  // Number of smallest set of smallest rings
}

// RenderResponse contains rendered image data.
// @Description Response containing the rendered image. For PNG/PDF the image field is base64-encoded; for SVG it is raw XML.
type RenderResponse struct {
	Image  string `json:"image"`                    // Base64-encoded image data (or raw SVG string)
	Format string `json:"format" example:"png"`     // Image format: png, svg, pdf
}

// CheckResponse contains structure validation results.
// @Description Response from structure validation.
type CheckResponse struct {
	Valid  bool     `json:"valid" example:"true"`                  // Whether the structure is valid
	Errors []string `json:"errors,omitempty"`                      // Validation error messages
	Struct string   `json:"struct,omitempty" example:"CCO"`        // Canonical representation of the structure
	Format string   `json:"format,omitempty" example:"smiles"`     // Format of the canonical representation
}

// InfoResponse contains library version information.
// @Description Response containing Indigo library and service version info.
type InfoResponse struct {
	IndigoVersion  string `json:"indigo_version" example:"1.x"`             // Indigo C library version
	ServiceName    string `json:"service_name" example:"go-indigo-service"` // Service name
	ServiceVersion string `json:"service_version" example:"1.0.0"`         // Service version
}

// ErrorResponse contains error information.
// @Description Error response returned when a request fails.
type ErrorResponse struct {
	Error   string `json:"error" example:"struct field is required"`   // Error message
	Details string `json:"details,omitempty" example:""`               // Detailed error information
}
