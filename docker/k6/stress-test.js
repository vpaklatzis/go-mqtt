import http from 'k6/http';

export const options = {
  // Key configurations for Stress in this section
  stages: [
    { duration: '1m', target: 200 }, // traffic ramp-up from 1 to a higher 200 users over 10 minutes.
    { duration: '1m', target: 200 }, // stay at higher 200 users for 10 minutes
    { duration: '1m', target: 0 }, // ramp-down to 0 users
  ],
};

export default () => {
  const urlRes = http.get('http://localhost:8080/api/v1/health');
};