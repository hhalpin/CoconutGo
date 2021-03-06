# BLS381 curve constants

from constants import *

SHA = 'sha256'   # hash type to use with this curve
EFS = 48   # Elliptic curve Field Size in bytes
CurveType = WEIERSTRASS

SexticTwist =  M_TYPE
SignOfX = NEGATIVEX
PairingFriendly = BLS

x=0xd201000000010000

if SignOfX == NEGATIVEX :
	p=(x**6+2*x**5-2*x**3-x+1)//3
	t=-x+1
else :
	p=(x**6-2*x**5+2*x**3+x+1)//3
	t=x+1

r=x*x*x*x-x*x+1

# elliptic curve
A=0
B=4

# generator point on G1
Gx=0x17F1D3A73197D7942695638C4FA9AC0FC3688C4F9774B905A14E3A3F171BAC586C55E83FF97A1AEFFB3AF00ADB22C6BB
Gy=0x8B3F481E3AAA0F1A09E30ED741D8AE4FCF5E095D5D00AF600DB18CB2C04B3EDD03CC744A2888AE40CAA232946C5E7E1

# Frobenius constant
Fra=0x1904D3BF02BB0667C231BEB4202C0D1F0FD603FD3CBD5F4F7B2443D784BAB9C4F67EA53D63E7813D8D0775ED92235FB8
Frb=0xFC3E2B36C4E03288E9E902231F9FB854A14787B6C7B36FEC0C8EC971F63C5F282D5AC14D6C7EC22CF78A126DDC4AF3

# Generator point on G2
Pxa=0x24AA2B2F08F0A91260805272DC51051C6E47AD4FA403B02B4510B647AE3D1770BAC0326A805BBEFD48056C8C121BDB8
Pxb=0x13E02B6052719F607DACD3A088274F65596BD0D09920B61AB5DA61BBDC7F5049334CF11213945D57E5AC7D055D042B7E
Pya=0xCE5D527727D6E118CC9CDC6DA2E351AADFD9BAA8CBDD3A76D429A695160D12C923AC9CC3BACA289E193548608B82801
Pyb=0x606C4A02EA734CC32ACD2B02BC28B99CB3E287E85A763AF267492AB572E99AB3F370D275CEC1DA1AAA9075FF05F79BE


