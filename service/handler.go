// Package main provides REST API handlers for go-indigo chemistry toolkit
// coding=utf-8
// @Project : go-indigo
// @Time    : 2026/03/16
// @Author  : chengxiang.luo
// @Email   : chengxiang.luo@foxmail.com
// @File    : handler.go
package main

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/cx-luo/go-indigo/core"
	"github.com/cx-luo/go-indigo/render"
	"github.com/gin-gonic/gin"
)

const serviceVersion = "1.0.0"

// IndigoService holds the shared Indigo session pool and related resources
type IndigoService struct {
	Pool *core.SessionPool
}

// NewIndigoService creates a new IndigoService with a session pool of the given size
func NewIndigoService(poolSize int) *IndigoService {
	return &IndigoService{
		Pool: core.NewSessionPool(poolSize),
	}
}

// loadStructure loads a molecule or reaction from the request body using the given Indigo session.
// It attempts auto-detection when input_format is empty or "auto".
func (s *IndigoService) loadMolecule(indigo *core.Indigo, req *StructRequest) (*moleculeHandle, error) {
	mol, err := indigo.LoadMoleculeFromString(req.Struct)
	if err != nil {
		mol, err = indigo.LoadStructureFromString(req.Struct, "")
		if err != nil {
			return nil, err
		}
	}
	return &moleculeHandle{mol: mol}, nil
}

type moleculeHandle struct {
	mol interface {
		Close() error
		Aromatize() error
		Dearomatize() error
		Layout() error
		Clean2D() error
		Standardize() error
		ToSmiles() (string, error)
		ToCanonicalSmiles() (string, error)
		ToMolfile() (string, error)
		ToCML() (string, error)
		ToCDXML() (string, error)
		ToJSON() (string, error)
		GrossFormula() (string, error)
		MolecularFormula() (string, error)
		MolecularWeight() (float64, error)
		MostAbundantMass() (float64, error)
		MonoisotopicMass() (float64, error)
		MassComposition() (string, error)
		TPSA(includeSP bool) (float64, error)
		NumRotatableBonds() (int, error)
		CountAtoms() (int, error)
		CountBonds() (int, error)
		CountHeavyAtoms() (int, error)
		CountComponents() (int, error)
		CountSSSR() (int, error)
	}
}

// handleInfo returns library version info
// GET /v2/indigo/info
func (s *IndigoService) handleInfo(c *gin.Context) {
	c.JSON(http.StatusOK, InfoResponse{
		IndigoVersion:  "1.x", // Indigo C library version
		ServiceName:    "go-indigo-service",
		ServiceVersion: serviceVersion,
	})
}

// handleAromatize aromatizes the input structure
// POST /v2/indigo/aromatize
func (s *IndigoService) handleAromatize(c *gin.Context) {
	var req StructRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body", Details: err.Error()})
		return
	}

	if req.Struct == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "struct field is required"})
		return
	}

	indigo := s.Pool.Get()
	defer s.Pool.Put(indigo)

	mh, err := s.loadMolecule(indigo, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to load structure", Details: err.Error()})
		return
	}
	defer mh.mol.Close()

	if err := mh.mol.Aromatize(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "aromatize failed", Details: err.Error()})
		return
	}

	result, format, err := convertToOutputFormat(mh, req.Options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "conversion failed", Details: err.Error()})
		return
	}

	c.JSON(http.StatusOK, StructResponse{
		Struct: result,
		Format: format,
	})
}

// handleDearomatize dearomatizes the input structure
// POST /v2/indigo/dearomatize
func (s *IndigoService) handleDearomatize(c *gin.Context) {
	var req StructRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body", Details: err.Error()})
		return
	}

	if req.Struct == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "struct field is required"})
		return
	}

	indigo := s.Pool.Get()
	defer s.Pool.Put(indigo)

	mh, err := s.loadMolecule(indigo, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to load structure", Details: err.Error()})
		return
	}
	defer mh.mol.Close()

	if err := mh.mol.Dearomatize(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "dearomatize failed", Details: err.Error()})
		return
	}

	result, format, err := convertToOutputFormat(mh, req.Options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "conversion failed", Details: err.Error()})
		return
	}

	c.JSON(http.StatusOK, StructResponse{
		Struct: result,
		Format: format,
	})
}

// handleCalculate calculates molecular properties
// POST /v2/indigo/calculate
func (s *IndigoService) handleCalculate(c *gin.Context) {
	var req CalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body", Details: err.Error()})
		return
	}

	if req.Struct == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "struct field is required"})
		return
	}

	indigo := s.Pool.Get()
	defer s.Pool.Put(indigo)

	mh, err := s.loadMolecule(indigo, &req.StructRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to load structure", Details: err.Error()})
		return
	}
	defer mh.mol.Close()

	resp := CalculateResponse{}
	allProps := len(req.Properties) == 0
	propSet := make(map[string]bool)
	for _, p := range req.Properties {
		propSet[strings.ToLower(p)] = true
	}

	shouldCalc := func(name string) bool {
		return allProps || propSet[name]
	}

	if shouldCalc("molecular-weight") {
		if w, err := mh.mol.MolecularWeight(); err == nil {
			resp.MolecularWeight = &w
		}
	}
	if shouldCalc("most-abundant-mass") {
		if m, err := mh.mol.MostAbundantMass(); err == nil {
			resp.MostAbundantMass = &m
		}
	}
	if shouldCalc("monoisotopic-mass") {
		if m, err := mh.mol.MonoisotopicMass(); err == nil {
			resp.MonoisotopicMass = &m
		}
	}
	if shouldCalc("mass-composition") {
		if mc, err := mh.mol.MassComposition(); err == nil {
			resp.MassComposition = mc
		}
	}
	if shouldCalc("gross-formula") {
		if gf, err := mh.mol.GrossFormula(); err == nil {
			resp.GrossFormula = gf
		}
	}
	if shouldCalc("molecular-formula") {
		if mf, err := mh.mol.MolecularFormula(); err == nil {
			resp.MolecularFormula = mf
		}
	}
	if shouldCalc("tpsa") {
		if t, err := mh.mol.TPSA(false); err == nil {
			resp.TPSA = &t
		}
	}
	if shouldCalc("num-rotatable-bonds") {
		if n, err := mh.mol.NumRotatableBonds(); err == nil {
			resp.NumRotatableBonds = &n
		}
	}
	if shouldCalc("num-atoms") {
		if n, err := mh.mol.CountAtoms(); err == nil {
			resp.NumAtoms = &n
		}
	}
	if shouldCalc("num-bonds") {
		if n, err := mh.mol.CountBonds(); err == nil {
			resp.NumBonds = &n
		}
	}
	if shouldCalc("num-heavy-atoms") {
		if n, err := mh.mol.CountHeavyAtoms(); err == nil {
			resp.NumHeavyAtoms = &n
		}
	}
	if shouldCalc("num-components") {
		if n, err := mh.mol.CountComponents(); err == nil {
			resp.NumComponents = &n
		}
	}
	if shouldCalc("num-sssr") {
		if n, err := mh.mol.CountSSSR(); err == nil {
			resp.NumSSSR = &n
		}
	}

	c.JSON(http.StatusOK, resp)
}

// handleConvert converts a structure to a different format
// POST /v2/indigo/convert
func (s *IndigoService) handleConvert(c *gin.Context) {
	var req ConvertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body", Details: err.Error()})
		return
	}

	if req.Struct == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "struct field is required"})
		return
	}

	if req.OutputFormat == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "output_format field is required"})
		return
	}

	indigo := s.Pool.Get()
	defer s.Pool.Put(indigo)

	mh, err := s.loadMolecule(indigo, &req.StructRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to load structure", Details: err.Error()})
		return
	}
	defer mh.mol.Close()

	result, err := convertMolecule(mh, req.OutputFormat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "conversion failed", Details: err.Error()})
		return
	}

	c.JSON(http.StatusOK, StructResponse{
		Struct: result,
		Format: req.OutputFormat,
	})
}

// handleClean performs 2D coordinate cleanup
// POST /v2/indigo/clean
func (s *IndigoService) handleClean(c *gin.Context) {
	var req StructRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body", Details: err.Error()})
		return
	}

	if req.Struct == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "struct field is required"})
		return
	}

	indigo := s.Pool.Get()
	defer s.Pool.Put(indigo)

	mh, err := s.loadMolecule(indigo, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to load structure", Details: err.Error()})
		return
	}
	defer mh.mol.Close()

	if err := mh.mol.Layout(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "layout failed", Details: err.Error()})
		return
	}

	if err := mh.mol.Clean2D(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "clean2d failed", Details: err.Error()})
		return
	}

	molfile, err := mh.mol.ToMolfile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "conversion failed", Details: err.Error()})
		return
	}

	c.JSON(http.StatusOK, StructResponse{
		Struct: molfile,
		Format: "molfile",
	})
}

// handleRender renders a structure to an image
// POST /v2/indigo/render
func (s *IndigoService) handleRender(c *gin.Context) {
	var req RenderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body", Details: err.Error()})
		return
	}

	if req.Struct == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "struct field is required"})
		return
	}

	outputFmt := req.OutputFormat
	if outputFmt == "" {
		outputFmt = "png"
	}

	indigo := s.Pool.Get()
	defer s.Pool.Put(indigo)

	mol, err := indigo.LoadMoleculeFromString(req.Struct)
	if err != nil {
		mol, err = indigo.LoadStructureFromString(req.Struct, "")
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to load structure", Details: err.Error()})
			return
		}
	}
	defer mol.Close()

	renderer, err := indigo.InitRenderer()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to init renderer", Details: err.Error()})
		return
	}
	defer renderer.DisposeRenderer()

	renderer.Options = &render.RenderOptions{
		OutputFormat:      outputFmt,
		ImageWidth:        1600,
		ImageHeight:       1600,
		BackgroundColor:   "1.0, 1.0, 1.0",
		BondLength:        40,
		RelativeThickness: 1.0,
		Margins:           "10, 10",
		StereoStyle:       "ext",
		LabelMode:         "hetero",
	}

	if req.Width > 0 {
		renderer.Options.ImageWidth = req.Width
	}
	if req.Height > 0 {
		renderer.Options.ImageHeight = req.Height
	}

	for k, v := range req.Options {
		_ = renderer.SetRenderOption(k, v)
	}

	if err := renderer.Apply(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to apply render options", Details: err.Error()})
		return
	}

	bufHandle, err := indigo.CreateWriteBuffer()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create buffer", Details: err.Error()})
		return
	}
	defer indigo.FreeObject(bufHandle)

	if err := renderer.Render(mol.Handle, bufHandle); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "render failed", Details: err.Error()})
		return
	}

	data, err := indigo.GetBufferData(bufHandle)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read render output", Details: err.Error()})
		return
	}

	if outputFmt == "svg" {
		c.JSON(http.StatusOK, RenderResponse{
			Image:  string(data),
			Format: outputFmt,
		})
		return
	}

	c.JSON(http.StatusOK, RenderResponse{
		Image:  base64.StdEncoding.EncodeToString(data),
		Format: outputFmt,
	})
}

// handleCheck validates the input structure
// POST /v2/indigo/check
func (s *IndigoService) handleCheck(c *gin.Context) {
	var req CheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body", Details: err.Error()})
		return
	}

	if req.Struct == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "struct field is required"})
		return
	}

	indigo := s.Pool.Get()
	defer s.Pool.Put(indigo)

	mol, err := indigo.LoadMoleculeFromString(req.Struct)
	if err != nil {
		mol, err = indigo.LoadStructureFromString(req.Struct, "")
		if err != nil {
			c.JSON(http.StatusOK, CheckResponse{
				Valid:  false,
				Errors: []string{err.Error()},
			})
			return
		}
	}
	defer mol.Close()

	smiles, _ := mol.ToCanonicalSmiles()
	c.JSON(http.StatusOK, CheckResponse{
		Valid:  true,
		Struct: smiles,
		Format: "smiles",
	})
}

// --- Helper functions ---

// convertToOutputFormat converts a molecule to the preferred output format.
// Defaults to canonical SMILES unless options["output_format"] is specified.
func convertToOutputFormat(mh *moleculeHandle, options map[string]string) (string, string, error) {
	format := "smiles"
	if f, ok := options["output_format"]; ok && f != "" {
		format = f
	}
	result, err := convertMolecule(mh, format)
	if err != nil {
		return "", "", err
	}
	return result, format, nil
}

// convertMolecule converts a molecule to the specified format
func convertMolecule(mh *moleculeHandle, format string) (string, error) {
	switch strings.ToLower(format) {
	case "smiles":
		return mh.mol.ToCanonicalSmiles()
	case "molfile", "mol", "sdf":
		return mh.mol.ToMolfile()
	case "cml":
		return mh.mol.ToCML()
	case "cdxml":
		return mh.mol.ToCDXML()
	case "json", "ket":
		return mh.mol.ToJSON()
	default:
		return mh.mol.ToCanonicalSmiles()
	}
}
