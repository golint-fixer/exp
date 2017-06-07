[BITS 32]

;global _start:function

global div_r8:function
global div_m8:function
global div_r16:function
global div_m16:function
global div_r32:function
global div_m32:function

section .text

;_start:
;	call    div_m8
;	; sys_exit(eax)
;	mov     ebx, eax
;	mov     eax, 1
;	int     0x80

;add:
;	mov     eax, 29
;	add     eax, 13
;	ret

;sub:
;	mov     eax, 55
;	sub     eax, 13
;	ret

;mul:
;	mov     eax, 2
;	mov     ebx, 21
;	mul     eax
;	ret

;imul:
;	mov     eax, 2
;	imul    eax, 21
;	ret

; === [ div arg ] ==============================================================
;
; --- [ 8-bit divisor arg ] ----------------------------------------------------
;
;           AX
;    ___________________ => AL (quotient) and AH (remainder)
;    arg (8-bit divisor)
;
div_r8:
	; 42 = 84 / 2
	mov     ax, 84
	mov     bl, 2
	div     bl
	and     eax, 0x000000FF
	ret

div_m8:
	; 42 = 84 / 2
	mov     ax, 84
	mov     byte [m8], 2
	div     byte [m8]
	and     eax, 0x000000FF
	ret

; --- [ 16-bit divisor arg ] ---------------------------------------------------
;
;           DX:AX
;    ____________________ => AX (quotient) and DX (remainder)
;    arg (16-bit divisor)
;
div_r16:
	; 42 = 84 / 2
	mov     dx, 0
	mov     ax, 84
	mov     bx, 2
	div     bx
	and     eax, 0x0000FFFF
	ret

div_m16:
	; 42 = 84 / 2
	mov     dx, 0
	mov     ax, 84
	mov     word [m16], 2
	div     word [m16]
	and     eax, 0x0000FFFF
	ret

; --- [ 32-bit divisor arg ] ---------------------------------------------------
;
;         EDX:EAX
;    ____________________ => EAX (quotient) and EDX (remainder)
;    arg (32-bit divisor)
;
div_r32:
	; 42 = 84 / 2
	mov     edx, 0
	mov     eax, 84
	mov     ebx, 2
	div     ebx
	ret

div_m32:
	; 42 = 84 / 2
	mov     edx, 0
	mov     eax, 84
	mov     dword [m32], 2
	div     dword [m32]
	ret

;imul:
;	mov     eax, 2
;	imul    eax, 21
;	ret

section .bss

; 8-bit memory variable.
m8: resb 1

; 16-bit memory variable.
m16: resw 1

; 32-bit memory variable.
m32: resd 1
