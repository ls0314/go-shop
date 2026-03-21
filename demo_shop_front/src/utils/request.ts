import axios from "axios";

const request = axios.create({
    baseURL: "http://localhost:9000/api/v1",
    timeout: 5000,
});

request.interceptors.request.use((config) => {
    const token = localStorage.getItem("access_token");
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

request.interceptors.response.use(
    (res) => {
        return res.data;
    },
    (err) => {
        return Promise.reject(err);
    }
);

export default request;