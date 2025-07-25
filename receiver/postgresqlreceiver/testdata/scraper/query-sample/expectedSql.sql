SELECT   
    COALESCE(datname, '') AS datname,
    COALESCE(usename, '') AS usename,
    COALESCE(client_addr::TEXT, '') AS client_addr,
    COALESCE(client_hostname, '') AS client_hostname,
    COALESCE(client_port::TEXT, '') AS client_port,
    COALESCE(query_start::TEXT, '') AS query_start,
    COALESCE(wait_event_type, '') AS wait_event_type,
    COALESCE(wait_event, '') AS wait_event,
    COALESCE(query_id::TEXT, '') AS query_id,
    COALESCE(pid::TEXT, '') AS pid,
    COALESCE(application_name::TEXT, '') AS application_name,
    EXTRACT(EPOCH FROM query_start) AS _query_start_timestamp,
    state,
    query
FROM pg_stat_activity
WHERE     
    coalesce(
      TRIM(query), 
      ''
    ) != ''
    AND NOT (

      query_start < TO_TIMESTAMP(123440.111)
      AND state = 'idle'
    )   
LIMIT 30;

