package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/mewrev/pe"
	"github.com/pkg/errors"
)

// dumpHeader dumps the PE header of the executable.
func dumpHeader(binPath string) error {
	buf := &bytes.Buffer{}
	file, err := pe.Open(binPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer file.Close()
	dosHdr, err := file.DOSHeader()
	if err != nil {
		return errors.WithStack(err)
	}
	dosStub, err := file.DOSStub()
	if err != nil {
		return errors.WithStack(err)
	}
	fileHdr, err := file.FileHeader()
	if err != nil {
		return errors.WithStack(err)
	}
	optHdr, err := file.OptHeader()
	if err != nil {
		return errors.WithStack(err)
	}
	sectHdrs, err := file.SectHeaders()
	if err != nil {
		return errors.WithStack(err)
	}
	_ = dosHdr
	_ = dosStub
	_ = fileHdr
	_ = optHdr
	_ = sectHdrs
	// Dump PE pre-header.
	//
	// ; PE header
	// ;
	// ;    file offset:    0x00000000
	// ;    virtual offset: 0x00400000
	const pePreHdrFormat = `
; PE header
;
;    file offset:    0x00000000
;    virtual offset: 0x%08X

SECTION hdr

`
	fmt.Fprintf(buf, pePreHdrFormat[1:], optHdr.ImageBase)

	// Dump DOS header.
	//
	//    ; === [ IMAGE_DOS_HEADER ] =====================================================
	//    mz_hdr:                                                         ; IMAGE_DOS_HEADER
	//                            dw      "MZ"                            ;    e_magic                            (Mark Zbikowski)
	//                            dw      0x0090                          ;    e_cblp
	//                            dw      0x0003                          ;    e_cp
	//                            dw      UNUSED                          ;    e_crlc
	//                            dw      0x0004                          ;    e_cparhdr
	//                            dw      UNUSED                          ;    e_minalloc
	//                            dw      0xFFFF                          ;    e_maxalloc
	//                            dw      UNUSED                          ;    e_ss
	//                            dw      0x00B8                          ;    e_sp
	//                            dw      UNUSED                          ;    e_csum
	//                            dw      UNUSED                          ;    e_ip
	//                            dw      UNUSED                          ;    e_cs
	//                            dw      0x0040                          ;    e_lfarlc
	//                            dw      UNUSED                          ;    e_ovno
	//            times 4         dw      UNUSED                          ;    e_res[4]
	//                            dw      UNUSED                          ;    e_oemid
	//                            dw      UNUSED                          ;    e_oeminfo
	//            times 10        dw      UNUSED                          ;    e_res2[10]
	//                            dd      pe_hdr - hdr_vstart             ;    e_lfanew
	//    ; === [/ IMAGE_DOS_HEADER ] ====================================================
	const dosHdrFormat = `
; === [ IMAGE_DOS_HEADER ] =====================================================
mz_hdr:                                                         ; IMAGE_DOS_HEADER
                        dw      "MZ"                            ;    e_magic                            (Mark Zbikowski)
                        dw      0x%04X                          ;    e_cblp
                        dw      0x%04X                          ;    e_cp
                        dw      0x%04X                          ;    e_crlc
                        dw      0x%04X                          ;    e_cparhdr
                        dw      0x%04X                          ;    e_minalloc
                        dw      0x%04X                          ;    e_maxalloc
                        dw      0x%04X                          ;    e_ss
                        dw      0x%04X                          ;    e_sp
                        dw      0x%04X                          ;    e_csum
                        dw      0x%04X                          ;    e_ip
                        dw      0x%04X                          ;    e_cs
                        dw      0x%04X                          ;    e_lfarlc
                        dw      0x%04X                          ;    e_ovno
        times 4         dw      0x0000                          ;    e_res[4]
                        dw      0x%04X                          ;    e_oemid
                        dw      0x%04X                          ;    e_oeminfo
        times 10        dw      0x0000                          ;    e_res2[10]
                        dd      pe_hdr - hdr_vstart             ;    e_lfanew
; === [/ IMAGE_DOS_HEADER ] ====================================================

`
	fmt.Fprintf(buf, dosHdrFormat[1:], dosHdr.LastPageSize, dosHdr.NPage, dosHdr.NReloc, dosHdr.NHdrPar, dosHdr.MinAlloc, dosHdr.MaxAlloc, dosHdr.SS, dosHdr.SP, dosHdr.Checksum, dosHdr.IP, dosHdr.CS, dosHdr.RelocTblOffset, dosHdr.OverlayNum, dosHdr.OEMID, dosHdr.OEMInfo)

	// Dump DOS stub.
	buf.WriteString("; DOS stub\n")
	for i, b := range dosStub {
		if i%8 == 0 {
			buf.WriteString("\n        db      ")
		} else {
			buf.WriteString(", ")
		}
		fmt.Fprintf(buf, "0x%02X", b)
	}
	buf.WriteString("\n")
	fmt.Println(hex.Dump(dosStub))

	// Dump PE header.
	//
	//    ; === [ IMAGE_NT_HEADERS ] =====================================================
	//    pe_hdr:                                                         ; IMAGE_NT_HEADERS
	//                            dd      "PE"                            ;    Signature                          (Portable Executable)
	//
	//    ; ------ [ IMAGE_FILE_HEADER ] -------------------------------------------------
	//    coff_hdr:                                                       ;    IMAGE_FILE_HEADER
	//                            dw      0x014C                          ;       Machine                         (x86)
	//                            dw      sect_hdr_count                  ;       NumberOfSections
	//                            dd      0x3B05AC00                      ;       TimeDateStamp                   (2001-05-19 - 01:10:56)
	//                            dd      0x00000000                      ;       PointerToSymbolTable
	//                            dd      0x00000000                      ;       NumberOfSymbols
	//                            dw      opt_hdr_size                    ;       SizeOfOptionalHeader
	//                            dw      0x010F                          ;       Characteristics                 (no local symbols, no line numbers, no relocations, exec, 32 bit)
	//    ; ------ [/ IMAGE_FILE_HEADER ] ------------------------------------------------
	//
	//    ; ------ [ IMAGE_OPTIONAL_HEADER ] ---------------------------------------------
	//    opt_hdr:                                                        ;    IMAGE_OPTIONAL_HEADER
	//
	//       file_align           equ     0x200
	//       sect_align           equ     0x1000                          ; (minimum section alignment: 4 KB)
	//
	//    %define round(n, r)     (((n + (r - 1)) / r) * r)

	const peHdrFormat = `
; === [ IMAGE_NT_HEADERS ] =====================================================
pe_hdr:                                                         ; IMAGE_NT_HEADERS
                        dd      "PE"                            ;    Signature                          (Portable Executable)

; ------ [ IMAGE_FILE_HEADER ] -------------------------------------------------
coff_hdr:                                                       ;    IMAGE_FILE_HEADER
                        dw      0x%04X                          ;       Machine                         (%s)
                        dw      sect_hdr_count                  ;       NumberOfSections
                        dd      0x%08X                      ;       TimeDateStamp                   (%s)
                        dd      0x%08X                      ;       PointerToSymbolTable
                        dd      0x%08X                      ;       NumberOfSymbols
                        dw      opt_hdr_size                    ;       SizeOfOptionalHeader
                        dw      0x%04X                          ;       Characteristics                 (%s)
; ------ [/ IMAGE_FILE_HEADER ] ------------------------------------------------
`
	fmt.Fprintf(buf, peHdrFormat, uint16(fileHdr.Arch), fileHdr.Arch, uint32(fileHdr.Created), fileHdr.Created, fileHdr.SymTblOffset, fileHdr.NSymbol, uint16(fileHdr.Flags), fileHdr.Flags)

	// Dump optional header.
	const optHdrFormat = `
; ------ [ IMAGE_OPTIONAL_HEADER ] ---------------------------------------------
opt_hdr:                                                        ;    IMAGE_OPTIONAL_HEADER

   file_align           equ     0x%04X
   sect_align           equ     0x%04X                          ; (minimum section alignment: %d KB)

%%define round(n, r)     (((n + (r - 1)) / r) * r)

; ___ [ standard fields ] ______________________________________________________
                        dw      0x%04X                          ;       Magic                           (%s)
                        db      0x%02X                            ;       MajorLinkerVersion
                        db      0x%02X                            ;       MinorLinkerVersion
                        dd      _text_size                      ;       SizeOfCode
                        dd      round(_rdata_size + _data_vsize + _rsrc_size, file_align) ;       SizeOfInitializedData
                        dd      0x%08X                      ;       SizeOfUninitializedData
                        dd      start - IMAGE_BASE              ;       AddressOfEntryPoint
                        dd      CODE_BASE                       ;       BaseOfCode
                        dd      DATA_BASE                       ;       BaseOfData

; ___ [ Windows-specific fields ] ______________________________________________
                        dd      IMAGE_BASE                      ;       ImageBase
                        dd      sect_align                      ;       SectionAlignment
                        dd      file_align                      ;       FileAlignment
                        dw      0x%04X                          ;       MajorOperatingSystemVersion
                        dw      0x%04X                          ;       MinorOperatingSystemVersion
                        dw      0x%04X                          ;       MajorImageVersion
                        dw      0x%04X                          ;       MinorImageVersion
                        dw      0x%04X                          ;       MajorSubsystemVersion
                        dw      0x%04X                          ;       MinorSubsystemVersion
                        dd      0x00000000                      ;       Win32VersionValue
                        dd      round(hdr_size, sect_align) + round(_text_vsize, sect_align) + round(_rdata_vsize, sect_align) + round(_data_vsize, sect_align) + round(_rsrc_vsize, sect_align) ;       SizeOfImage
                        dd      round(hdr_size, file_align)     ;       SizeOfHeaders
                        dd      0x%08X                      ;       CheckSum
                        dw      0x%04X                          ;       Subsystem                       (%s)
                        dw      0x%04X                          ;       DllCharacteristics              (%s)
                        dd      0x%08X                      ;       SizeOfStackReserve
                        dd      0x%08X                      ;       SizeOfStackCommit
                        dd      0x%08X                      ;       SizeOfHeapReserve
                        dd      0x%08X                      ;       SizeOfHeapCommit
                        dd      0x%08X                      ;       LoaderFlags
                        dd      data_dir_count                  ;       NumberOfRvaAndSizes
`
	fmt.Fprintf(buf, optHdrFormat, optHdr.FileAlign, optHdr.SectAlign, optHdr.SectAlign/1024, uint16(optHdr.State), optHdr.State, optHdr.MajorLinkVer, optHdr.MinorLinkVer, optHdr.BSSSize, optHdr.MajorOSVer, optHdr.MinorOSVer, optHdr.MajorImageVer, optHdr.MinorImageVer, optHdr.MajorSubsystemVer, optHdr.MinorSubsystemVer, optHdr.Checksum, uint16(optHdr.Subsystem), optHdr.Subsystem, uint16(optHdr.Flags), optHdr.Flags, optHdr.ReserveStackSize, optHdr.InitStackSize, optHdr.ReserveHeapSize, optHdr.InitHeapSize, optHdr.LoaderFlags)

	// Store output.
	outPath := filepath.Join(outDir, "pe-hdr.asm")
	dbg.Printf("creating %q\n", outPath)
	if err := ioutil.WriteFile(outPath, buf.Bytes(), 0644); err != nil {
		return errors.WithStack(err)
	}
	return nil
}