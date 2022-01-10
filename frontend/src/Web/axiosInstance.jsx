/**
 * SecureRequest.jsx
 * A wrapper for secure requests between backend and frontend.
 *
 * Author: 190014935
 */

import axios from 'axios'

// Axios instance used for making needed requests.
const axiosInstance = axios.create({
  baseURL: process.env.BACKEND_ADDRESS,
  timeout: 1000,
  headers: {"X-FOREIGNJOURNAL-SECURITY-TOKEN": process.env.BACKEND_TOKEN,
            "Content-Type": "application/json"}
});

export default axiosInstance;
