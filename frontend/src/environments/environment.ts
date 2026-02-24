const buildApiBaseUrl = '__API_BASE_URL__';
if (buildApiBaseUrl === '__API_BASE_URL__') {
  throw new Error('API_BASE_URL is not defined at build time.');
}

export const environment = {
  apiBaseUrl: buildApiBaseUrl,
};
