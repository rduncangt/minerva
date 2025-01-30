# To-Do List

## Documentation

- [ ] Update and expand `README.md` with:
  - Project overview.
  - Usage instructions.
  - Dependencies and installation steps.

---

## Testing and Validation

- [ ] Write unit tests for all new and existing functionality.
- [ ] Ensure integration tests simulate end-to-end workflows with:
  - Mock data for log inputs.
  - A test database environment.
- [ ] Validate edge cases such as:
  - Invalid or missing log fields.
  - Transient database connection failures.

---

## Performance Optimization

- [ ] Improve database query efficiency, particularly for duplicate checks.
- [ ] Implement batch processing for database inserts.
- [ ] Investigate caching mechanisms for previously processed data.
- [ ] Benchmark the pipeline to identify and resolve bottlenecks.

---

## System Enhancements

- [ ] Add error handling and logging for:
  - Geolocation API failures.
  - Database insert errors.
  - File or stream input issues.
- [ ] Implement rate-limiting for external API calls.
- [ ] Research and integrate additional data sources for geolocation or threat intelligence.
- [ ] Add First Seen/Last Seen Timestamps to IP address intel
- [ ] Integrate with threat intelligence APIs, AbuseIPDB, VirusTotal.
  - Implement threat_score and/or malicious_flags column in table
  
---

## Future Goals

- [ ] Modularize the processing pipeline for better maintainability.
- [ ] Explore visualization tools for presenting insights, such as:
  - Attack trends and geographic data mapping.
  - Frequent source IPs and targeted ports.
- [ ] Design database schema optimizations for querying and analysis.
