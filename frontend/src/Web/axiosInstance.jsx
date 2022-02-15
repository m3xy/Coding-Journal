/**
 * SecureRequest.jsx
 * A wrapper for secure requests between backend and frontend.
 *
 * Author: 190014935
 */

import axios from "axios";
import JwtService from "./jwt.service";

// Axios instance used for making needed requests.
const axiosInstance = axios.create({
  baseURL: process.env.BACKEND_ADDRESS,
  timeout: 1000,
  headers: {
    "Content-Type": "application/json",
  },
});

// Handler for 401
axiosInstance.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    var origRequest = error.config;
    if (origRequest.url !== "/token" && error.response) {
      if (error.response.status === 401 && !origRequest._retry) {
        origRequest._retry = true;
        try {
          const rs = await axiosInstance({
            method: "get",
            url: "/auth/token",
            headers: {
              refresh_token: "Refresh " + JwtService.getRefreshToken(),
            },
          });
          JwtService.updateAccessToken(rs.data.content.token);
          return axiosInstance(origRequest);
        } catch (err) {
          return Promise.reject(error);
        }
      }
    }
    return Promise.reject(error);
  }
);

axiosInstance.interceptors.request.use(
  (config) => {
    const token = JwtService.getAccessToken();
    if (token) {
      config.headers["bearer_token"] = "Bearer " + token;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

export default axiosInstance;
