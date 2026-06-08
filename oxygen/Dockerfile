FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS base
USER root
RUN apt-get update && apt-get install -y --no-install-recommends curl && rm -rf /var/lib/apt/lists/*
USER $APP_UID
WORKDIR /app
EXPOSE 8082

FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
ARG BUILD_CONFIGURATION=Release
WORKDIR /src
COPY ["Oxygen.API/Oxygen.API.csproj", "Oxygen.API/"]
RUN dotnet restore "Oxygen.API/Oxygen.API.csproj"
COPY . .
WORKDIR "/src/Oxygen.API"
RUN dotnet build "Oxygen.API.csproj" -c $BUILD_CONFIGURATION -o /app/build

FROM build AS publish
ARG BUILD_CONFIGURATION=Release
RUN dotnet publish "Oxygen.API.csproj" -c $BUILD_CONFIGURATION -o /app/publish /p:UseAppHost=false

FROM base AS final
WORKDIR /app
COPY --from=publish /app/publish .
ENV ASPNETCORE_URLS=http://+:8082
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD curl -f http://localhost:8082/loan/health || exit 1
ENTRYPOINT ["dotnet", "Oxygen.API.dll"]
