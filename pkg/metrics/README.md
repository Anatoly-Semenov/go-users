# Go-Users Metrics Documentation

## Overview

The Go-Users application now includes a comprehensive metrics system that provides detailed insights into application performance, security, and system health. This document describes the available metrics, monitoring setup, and how to use the system.

## Metrics Categories

### 1. Business Metrics

#### User Metrics
- `users_total_count` - Total number of users in the system
- `users_registered_total` - Counter of user registrations
- `users_active_daily` - Number of active users in the last 24 hours
- `users_active_weekly` - Number of active users in the last week
- `users_by_role{role}` - Number of users by role (admin/user)
- `user_operations_total{operation, status}` - User operations (create/update/delete/get/list)
- `user_login_attempts_total{status}` - Login attempts (success/failed)
- `user_login_success_rate` - Login success rate percentage

#### Authentication Metrics
- `jwt_tokens_issued_total` - Total JWT tokens issued
- `jwt_tokens_validated_total{status}` - Token validations (valid/invalid/expired)
- `jwt_token_validation_duration_seconds` - Time spent validating tokens
- `password_reset_requests_total` - Password reset requests

### 2. Security Metrics

#### IP Blocking & Brute Force Protection
- `ip_blocks_total{type}` - Total IP blocks (temporary/permanent)
- `ip_blocks_active{type}` - Active IP blocks
- `ip_blocks_created_total{type, reason}` - Created IP blocks
- `ip_blocks_expired_total` - Expired IP blocks
- `bruteforce_attempts_total{ip}` - Brute force attempts by IP
- `bruteforce_detected_total` - Detected brute force attacks
- `failed_login_attempts_total{ip}` - Failed login attempts by IP
- `blocked_requests_total{reason}` - Blocked requests
- `suspicious_activity_detected_total{type}` - Suspicious activities
- `security_violations_total{type}` - Security violations

### 3. API Metrics

#### HTTP API
- `http_requests_total{method, endpoint, status}` - HTTP requests
- `http_request_duration_seconds{method, endpoint, status}` - Request duration
- `http_request_size_bytes{method, endpoint}` - Request size
- `http_response_size_bytes{method, endpoint}` - Response size
- `http_concurrent_requests` - Concurrent HTTP requests
- `http_login_duration_seconds` - Login request duration
- `http_registration_duration_seconds` - Registration request duration
- `http_user_lookup_duration_seconds` - User lookup duration

#### gRPC API
- `grpc_requests_total{method, status}` - gRPC requests
- `grpc_request_duration_seconds{method, status}` - Request duration
- `grpc_request_size_bytes{method}` - Request size
- `grpc_response_size_bytes{method}` - Response size
- `grpc_concurrent_requests` - Concurrent gRPC requests
- `grpc_stream_messages_sent_total` - Stream messages sent
- `grpc_stream_messages_received_total` - Stream messages received

### 4. Infrastructure Metrics

#### Database (PostgreSQL)
- `db_connections_active` - Active database connections
- `db_connections_idle` - Idle database connections
- `db_connections_max` - Maximum database connections
- `db_connection_wait_duration_seconds` - Connection wait time
- `db_operations_total{operation, table}` - Database operations
- `db_operation_duration_seconds{operation, table}` - Operation duration
- `db_rows_affected_total{operation, table}` - Rows affected
- `db_query_errors_total{error_type}` - Database errors
- `db_user_queries_total{type}` - User-related queries

#### Redis
- `redis_connections_active` - Active Redis connections
- `redis_commands_total{command}` - Redis commands executed
- `redis_command_duration_seconds{command}` - Command duration
- `redis_memory_used_bytes` - Redis memory usage
- `redis_keys_total{type}` - Redis keys by type
- `redis_ip_blocks_total` - IP blocks in Redis
- `redis_ip_block_hits_total` - Cache hits
- `redis_ip_block_misses_total` - Cache misses

### 5. System Metrics

#### Application Performance
- `app_uptime_seconds` - Application uptime
- `app_restart_total` - Application restarts
- `app_health_check_duration_seconds` - Health check duration
- `middleware_request_duration_seconds{middleware}` - Middleware processing time
- `panic_recoveries_total` - Panic recoveries

#### Error Metrics
- `errors_total{component, error_type}` - Errors by component
- `error_rate{component}` - Error rate by component
- `validation_errors_total{field}` - Validation errors
- `auth_errors_total{type}` - Authentication errors

### 6. SLA/SLO Metrics
- `service_availability_percentage` - Service availability
- `endpoint_availability{endpoint}` - Endpoint availability
- `response_time_percentiles{percentile}` - Response time percentiles
- `error_budget_remaining` - Remaining error budget

### 7. Alert Metrics
- `high_error_rate{threshold}` - High error rate indicator
- `slow_response_time{threshold}` - Slow response indicator
- `database_connection_exhaustion` - DB connection exhaustion
- `memory_usage_high{threshold}` - High memory usage
- `active_bruteforce_attacks` - Active brute force attacks
- `potential_ddos_detected` - Potential DDoS indicator
