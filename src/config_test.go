package main

import (
	"reflect"
	"testing"
)

func Test_decodeBitwardenDataJson1(t *testing.T) {
	type args struct {
		byteData []byte
	}
	tests := []struct {
		name    string
		args    args
		want    BwData
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "before-1.21.1",
			args: args{
				byteData: []byte(`{"installedVersion":"1.20.0","appId":"appid-1234","accessToken":"ThisIsTheaccessToken","refreshToken":"ThisIsRefreshToken","userEmail":"bitwarden@test.com","userId":"ThisIsUserId","kdf":0,"kdfIterations":50000,"keyHash":"ThisIsKeyHash","encKey":"ThisIsEncKey","encPrivateKey":"ThisIsEncPrivateKey","__PROTECTED__key":"ThisIsProtectedKey","encProviderKeys":{}}`),
			},
			wantErr: false,
			want: BwData{
				path:             "",
				InstalledVersion: "1.20.0",
				UserEmail:        "bitwarden@test.com",
				UserId:           "ThisIsUserId",
				ActiveUserId:     "",
				ProtectedKey:     "ThisIsProtectedKey",
				EncKey:           "ThisIsEncKey",
				Kdf:              0,
				KdfIterations:    50000,
				Global:           BwGlobalData{},
				Profile:          BwProfileData{},
				Keys:             BwKeyData{},
				Tokens:           BwTokens{},
				Unused:           nil,
			},
		},
		{
			name: "since-1.21.1",
			args: args{
				byteData: []byte(`{"global":{"installedVersion":"1.21.1"},"userIdBlaBlubb":{"keys":{"masterKeyEncryptedUserKey":"ThisIsCryptoSymmetricKeyEncrypted","privateKey":{"encrypted":"ThisIsPrivateKeyEncrypted"},"apiKeyClientSecret":"ThisIsApiKeyClientSecret","legacyEtmKey":null},"profile":{"userId":"userIdBlaBlubb","email":"bitwarden@test.com","kdfIterations":50000,"kdfType":0,"lastSync":"2022-02-28T16:58:54.900Z","everBeenUnlocked":true},"tokens":{"accessToken":"ThisIsAccessToken"}},"activeUserId":"userIdBlaBlubb","__PROTECTED__userIdBlaBlubb_user_auto":"ThisIs__Protected__masterkey"}`),
			},
			wantErr: false,
			want: BwData{
				path:             "",
				InstalledVersion: "1.21.1",
				UserEmail:        "bitwarden@test.com",
				UserId:           "userIdBlaBlubb",
				ActiveUserId:     "userIdBlaBlubb",
				ProtectedKey:     "ThisIs__Protected__masterkey",
				EncKey:           "ThisIsCryptoSymmetricKeyEncrypted",
				Kdf:              0,
				KdfIterations:    50000,
				Global: BwGlobalData{
					InstalledVersion: "1.21.1",
				},
				Profile: BwProfileData{
					EverBeenUnlocked: true,
					LastSync:         "2022-02-28T16:58:54.900Z",
					KdfIterations:    50000,
					KdfType:          0,
					Email:            "bitwarden@test.com",
					UserId:           "userIdBlaBlubb",
				},
				Keys: BwKeyData{
					ApiKeyClientSecret: "ThisIsApiKeyClientSecret",
					CryptoSymmetricKey: BwCryptoSymmetricKey{
						Encrypted: "ThisIsCryptoSymmetricKeyEncrypted",
					},
					PrivateKey: BwPrivateKey{
						Encrypted: "ThisIsPrivateKeyEncrypted",
					},
				},
				Tokens: BwTokens{
					AccessToken: "ThisIsAccessToken",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeBitwardenDataJson(tt.args.byteData)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeBitwardenDataJson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeBitwardenDataJson() got = %v, want %v", got, tt.want)
			}
		})
	}
}
