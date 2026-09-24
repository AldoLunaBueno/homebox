export default defineNuxtRouteMiddleware(async to => {
  const ctx = useAuthContext();
  const api = useUserApi();
  const redirectTo = useState("authRedirect");

  if (!ctx.isAuthorized()) {
    if (to.path !== "/") {
      console.debug("[middleware/auth] isAuthorized returned false, redirecting to /");
      redirectTo.value = to.path;
      return navigateTo("/");
    }
  }

  if (!ctx.user) {
    console.log("Fetching user data");
    const { data, error } = await api.user.self();
    if (error) {
      if (to.path !== "/") {
        console.debug("[middleware/user] user is null and fetch failed, redirecting to /");
        redirectTo.value = to.path;
        return navigateTo("/");
      }
    }

    ctx.user = data.item;
  }

  // GUARDIÁN DE RUTAS NATIVO
  const protectedRoutes = ["/templates", "/maintenance", "/collection"];
  const isTryingToAccessProtectedRoute = protectedRoutes.some(route => to.path.startsWith(route));

  // Si intenta acceder a zona protegida y NO es superusuario en la BD, lo bloqueamos
  if (isTryingToAccessProtectedRoute && !ctx.user?.isSuperuser) {
    console.warn(`[Seguridad] Bloqueando acceso a operario hacia: ${to.path}`);
    return navigateTo("/home");
  }
});
