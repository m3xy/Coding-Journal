class JwtService {
  getRefreshToken() {
    return localStorage.getItem("refreshToken");
  }

  getAccessToken() {
    return sessionStorage.getItem("accessToken");
  }

  updateAccessToken(token) {
    sessionStorage.setItem("accessToken", token);
  }

  setUser(token, refresh_token) {
    sessionStorage.setItem("accessToken", token);
    localStorage.setItem("refreshToken", refresh_token);
  }

  rmUser() {
    sessionStorage.removeItem("accessToken");
    localStorage.removeItem("refreshToken");
  }
}

export default new JwtService();
