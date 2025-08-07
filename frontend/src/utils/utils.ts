export const isAuthenticated = async () => {
  try {
    const response = await fetch(`http://localhost:3000/auth/status`, {
      credentials: "include",
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    });

    return response.ok;
  } catch (error) {
    console.error(error);
    return false;
  }
}

export const isVerified = async () => {
  try {
    const response = await fetch(`http://localhost:3000/auth/verified`, {
      credentials: "include",
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    });

    return response.ok;
  } catch (error) {
    console.error(error);
    return false;
  }
}

export const getCSRFToken = (): string => {
        const name = "csrf_token=";
        const decoded = decodeURIComponent(document.cookie);
        const cookies = decoded.split(';');

        for (let cookie of cookies) {
            cookie = cookie.trim();
            if (cookie.startsWith(name)) {
                return cookie.substring(name.length);
            }
        }

        return "";
    }
