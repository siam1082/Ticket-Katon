import axios from 'axios';

const API = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// রিকোয়েস্ট পাঠানোর আগে লোকালস্টোরেজ থেকে টোকেন ইনজেক্ট করা
API.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export default API;