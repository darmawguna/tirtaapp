package services

import (
	"fmt"
	"strings"

	"github.com/darmawguna/tirtaapp.git/dto"
)

type complaintDef struct {
	Code  string
	Label string
}

var complaintCatalog = map[dto.ComplaintPhase][]complaintDef{
	dto.PhasePreHD: {
		{Code: "sesak_napas", Label: "Sesak napas"},
		{Code: "kaki_bengkak", Label: "Kaki bengkak"},
		{Code: "nyeri_perut", Label: "Nyeri perut"},
		{Code: "nyeri_kaki", Label: "Nyeri kaki"},
		{Code: "pusing", Label: "Pusing"},
		{Code: "lemas_cepat_lelah", Label: "Lemas / cepat lelah"},
		{Code: "gatal_gatal", Label: "Gatal-gatal"},
		{Code: "nafsu_makan_menurun", Label: "Nafsu makan menurun"},
		{Code: "bau_mulut", Label: "Bau mulut"},
		{Code: "kulit_kering_bersisik", Label: "Kulit kering / bersisik"},
		{Code: "kram_otot", Label: "Kram otot"},
		{Code: "sulit_tidur", Label: "Sulit tidur"},
		{Code: "gangguan_konsentrasi", Label: "Gangguan konsentrasi"},
		{Code: "hipertensi", Label: "Hipertensi"},
		{Code: "mual", Label: "Mual"},
		{Code: "sakit_kepala", Label: "Sakit kepala"},
		{Code: "keluhan_lainnya", Label: "Keluhan Lainnya"},
	},
	dto.PhasePostHD: {
		{Code: "pusing", Label: "Pusing"},
		{Code: "lemas", Label: "Lemas"},
		{Code: "kram_otot", Label: "Kram otot"},
		{Code: "mual_muntah", Label: "Mual / muntah"},
		{Code: "sakit_kepala", Label: "Sakit kepala"},
		{Code: "gatal", Label: "Gatal"},
		{Code: "demam_menggigil", Label: "Demam dan menggigil"},
		{Code: "pendarahan_akses", Label: "Pendarahan dari akses (AV Shunt / Double Lumen)"},
		{Code: "bengkak_nyeri_double_lumen", Label: "Bengkak atau nyeri pada double lumen"},
		{Code: "sesak_napas_berlanjut", Label: "Sesak napas berlanjut"},
		{Code: "nyeri_dada", Label: "Nyeri dada"},
		{Code: "kedinginan_ekstrem", Label: "Kedinginan ekstrem"},
		{Code: "nyeri_punggung_pinggang", Label: "Nyeri punggung atau pinggang"},
		{Code: "reaksi_alergi", Label: "Reaksi alergi (ruam, gatal, sesak)"},
		{Code: "keluhan_lainnya", Label: "Keluhan Lainnya"},
	},
}

func validatePhase(phase dto.ComplaintPhase) error {
	if phase != dto.PhasePreHD && phase != dto.PhasePostHD {
		return fmt.Errorf("invalid phase: %s", phase)
	}
	return nil
}

func catalogMaps(phase dto.ComplaintPhase) (map[string]string, error) {
	if err := validatePhase(phase); err != nil {
		return nil, err
	}
	defs := complaintCatalog[phase]
	m := make(map[string]string, len(defs))
	for _, d := range defs {
		m[d.Code] = d.Label
	}
	return m, nil
}

func normalizeCode(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}
