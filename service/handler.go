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

// loadMolecule loads a molecule from the request body using the given Indigo session.
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

// handleInfo godoc
//
//	@Summary		Get library version info
//	@Description	Returns the version of the Indigo C library and the go-indigo service
//	@Tags			indigo
//	@Produce		json
//	@Success		200	{object}	InfoResponse
//	@Router			/v2/indigo/info [get]
func (s *IndigoService) handleInfo(c *gin.Context) {
	c.JSON(http.StatusOK, InfoResponse{
		IndigoVersion:  "1.x",
		ServiceName:    "go-indigo-service",
		ServiceVersion: serviceVersion,
	})
}

// handleAromatize godoc
//
//	@Summary		Aromatize structure
//	@Description	Aromatize the input chemical structure (molecule or reaction)
//	@Tags			indigo
//	@Accept			json
//	@Produce		json
//	@Param			request	body		StructRequest	true	"Structure to aromatize"
//	@Success		200		{object}	StructResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/v2/indigo/aromatize [post]
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

// handleDearomatize godoc
//
//	@Summary		Dearomatize structure
//	@Description	Remove aromaticity from the input chemical structure
//	@Tags			indigo
//	@Accept			json
//	@Produce		json
//	@Param			request	body		StructRequest	true	"Structure to dearomatize"
//	@Success		200		{object}	StructResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/v2/indigo/dearomatize [post]
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

// handleCalculate godoc
//
//	@Summary		Calculate molecular properties
//	@Description	Calculate properties (molecular weight, formula, TPSA, etc.) for the input structure. If properties array is empty, all properties are computed.
//	@Tags			indigo
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CalculateRequest	true	"Structure and optional property list"
//	@Success		200		{object}	CalculateResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/v2/indigo/calculate [post]
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

	mh, err := s.loadMolecule(indigo, &StructRequest{
		Struct:      req.Struct,
		InputFormat: req.InputFormat,
		Options:     req.Options,
	})
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

// handleConvert godoc
//
//	@Summary		Convert structure format
//	@Description	Convert a chemical structure to a different format (SMILES, Molfile, CML, CDXML, JSON/KET)
//	@Tags			indigo
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ConvertRequest	true	"Structure and target output format"
//	@Success		200		{object}	StructResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/v2/indigo/convert [post]
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

	mh, err := s.loadMolecule(indigo, &StructRequest{
		Struct:      req.Struct,
		InputFormat: req.InputFormat,
		Options:     req.Options,
	})
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

// handleClean godoc
//
//	@Summary		Clean up structure coordinates
//	@Description	Perform 2D layout and coordinate cleanup for the input structure. Returns the result in Molfile format.
//	@Tags			indigo
//	@Accept			json
//	@Produce		json
//	@Param			request	body		StructRequest	true	"Structure to clean"
//	@Success		200		{object}	StructResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/v2/indigo/clean [post]
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

// handleRender godoc
//
//	@Summary		Render structure to image
//	@Description	Render a chemical structure to PNG, SVG, or PDF. For PNG/PDF the response image field is base64-encoded; for SVG it is raw XML.
//	@Tags			indigo
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RenderRequest	true	"Structure and render options"
//	@Success		200		{object}	RenderResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/v2/indigo/render [post]
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

// handleCheck godoc
//
//	@Summary		Validate structure
//	@Description	Verify whether the input string is a valid chemical structure (Molfile, SMILES, CML, InChI, etc.)
//	@Tags			indigo
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CheckRequest	true	"Structure to validate"
//	@Success		200		{object}	CheckResponse
//	@Failure		400		{object}	ErrorResponse
//	@Router			/v2/indigo/check [post]
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
