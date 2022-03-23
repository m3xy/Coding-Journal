/**
 * SecureRequest.jsx
 * A wrapper for secure requests between backend and frontend.
 *
 * Author: 190014935
 */

import axios from "axios"
import JwtService from "./jwt.service"

// Axios instance used for making needed requests.
const axiosInstance = axios.create({
	baseURL: process.env.BACKEND_ADDRESS,
	timeout: 4000,
	headers: {
		"Content-Type": "application/json"
	}
})

const TOKEN_ROUTE = "/auth/token"

// Handler for 401
axiosInstance.interceptors.response.use(
	(response) => {
		return response
	},
	async (error) => {
		var origRequest = error.config
		if (origRequest.url !== TOKEN_ROUTE && error.response) {
			if (error.response.status === 401 && !origRequest._retry) {
				origRequest._retry = true
				try {
					const rs = await axiosInstance({
						method: "get",
						url: TOKEN_ROUTE,
						headers: {
							refresh_token:
								"Refresh " + JwtService.getRefreshToken()
						}
					})
					JwtService.setUser(
						rs.data.access_token,
						rs.data.refresh_token
					)
					return axiosInstance(origRequest)
				} catch (err) {
					return Promise.reject(error)
				}
			}
		}
		return Promise.reject(error)
	}
)

axiosInstance.interceptors.request.use(
	(config) => {
		const token = JwtService.getAccessToken()
		if (token) {
			config.headers["bearer_token"] = "Bearer " + token
		}
		return config
	},
	(error) => {
		return Promise.reject(error)
	}
)

export default axiosInstance
