import jwt_decode from "jwt-decode"

class JwtService {
	getRefreshToken() {
		return localStorage.getItem("refreshToken")
	}

	getAccessToken() {
		return sessionStorage.getItem("accessToken")
	}

	setUser(token, refresh_token) {
		sessionStorage.setItem("accessToken", token)
		localStorage.setItem("refreshToken", refresh_token)
	}

	rmUser() {
		sessionStorage.removeItem("accessToken")
		localStorage.removeItem("refreshToken")
	}

	getUserID() {
		let token = this.getAccessToken()
		if (token) {
			return jwt_decode(token).userId
		} else return null
	}

	getUserType() {
		let token = this.getAccessToken()
		if (token) {
			return jwt_decode(token).userType
		} else return null
	}
}

export default new JwtService()
