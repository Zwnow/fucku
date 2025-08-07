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
