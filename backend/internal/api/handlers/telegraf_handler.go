package handlers

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/logging"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
)

// IngestTelegraf accepts Telegraf JSON payloads (single object or array) and
// maps them to internal metrics. This supports installing a Telegraf agent on
// hosts and sending system metrics to KineticOps.
func IngestTelegraf(c *fiber.Ctx) error {
	tidLoc := c.Locals("tenant_id")
	agentTokenUsed := false
	if at := c.Locals("agent_token"); at != nil {
		agentTokenUsed = true
	}

	body := c.Body()

	// Try to decode into a generic interface to support arrays or objects
	var raw interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		// If the payload isn't JSON array/object, try to parse as single line
		logging.Warnf("[TELEGRAF] invalid json payload: %v", err)
		return c.Status(400).JSON(fiber.Map{"error": "invalid json"})
	}

	processPoint := func(obj map[string]interface{}) {
		// Telegraf points often have 'measurement', 'tags', 'fields', 'time'
		fields := map[string]interface{}{}
		tags := map[string]interface{}{}

		if v, ok := obj["fields"]; ok {
			if m, ok2 := v.(map[string]interface{}); ok2 {
				fields = m
			}
		} else {
			// fallback: top-level fields
			fields = obj
		}
		if v, ok := obj["tags"]; ok {
			if m, ok2 := v.(map[string]interface{}); ok2 {
				tags = m
			}
		}

		// If tags contain host_id, prefer that for mapping to HostMetric
		var hostID int64
		if hid, ok := tags["host_id"]; ok {
			switch t := hid.(type) {
			case float64:
				hostID = int64(t)
			case int64:
				hostID = t
			case string:
				// ignore parse errors
			}
		}

		// If fields appear to be full host metrics, map to HostMetric
		// Recognize by presence of cpu, memory, disk related fields
		if hostID != 0 {
			// construct HostMetric from fields when possible
			hm := &services.HostMetric{}
			hm.HostID = hostID
			if v, ok := fields["cpu_usage"]; ok {
				if f, ok2 := v.(float64); ok2 {
					hm.CPUUsage = f
				}
			}
			if v, ok := fields["memory_usage"]; ok {
				if f, ok2 := v.(float64); ok2 {
					hm.MemoryUsage = f
				}
			}
			if v, ok := fields["memory_total"]; ok {
				if f, ok2 := v.(float64); ok2 {
					hm.MemoryTotal = f
				}
			}
			if v, ok := fields["memory_used"]; ok {
				if f, ok2 := v.(float64); ok2 {
					hm.MemoryUsed = f
				}
			}
			if v, ok := fields["disk_usage"]; ok {
				if f, ok2 := v.(float64); ok2 {
					hm.DiskUsage = f
				}
			}
			if v, ok := fields["disk_total"]; ok {
				if f, ok2 := v.(float64); ok2 {
					hm.DiskTotal = f
				}
			}
			if v, ok := fields["disk_used"]; ok {
				if f, ok2 := v.(float64); ok2 {
					hm.DiskUsed = f
				}
			}
			if v, ok := fields["network_in"]; ok {
				if f, ok2 := v.(float64); ok2 {
					hm.NetworkIn = f
				}
			}
			if v, ok := fields["network_out"]; ok {
				if f, ok2 := v.(float64); ok2 {
					hm.NetworkOut = f
				}
			}
			if v, ok := fields["uptime"]; ok {
				if f, ok2 := v.(float64); ok2 {
					hm.Uptime = int64(f)
				}
			}
			if v, ok := fields["load_average"]; ok {
				if s, ok2 := v.(string); ok2 {
					hm.LoadAverage = s
				}
			}

			// best-effort save via existing pipeline
			if err := services.SaveHostMetrics(hm); err != nil {
				logging.Errorf("[TELEGRAF] failed to save host metrics for host %d: %v", hostID, err)
			}
			return
		}

		// Otherwise, map each field to a metric and store
		for k, v := range fields {
			if f, ok := v.(float64); ok {
				// Try to find host id in tags
				var hid int64 = 0
				if t, ok2 := tags["host_id"]; ok2 {
					if tf, ok3 := t.(float64); ok3 {
						hid = int64(tf)
					}
				}
				// Skip if no host_id
				if hid == 0 {
					continue
				}

				// Resolve tenant id: prefer JWT tenant, otherwise look up host owner when agent token used
				var tenantID int64
				if tidLoc != nil {
					tenantID = tidLoc.(int64)
				} else if agentTokenUsed {
					// Try to lookup host owner
					if h, err := services.GetHostByID(hid, 0); err == nil && h != nil {
						tenantID = h.TenantID
					} else {
						logging.Warnf("[TELEGRAF] unable to resolve tenant for host %d: %v", hid, err)
						continue
					}
				} else {
					// No authenticated tenant
					continue
				}

				if err := services.CollectMetric(hid, tenantID, k, f, nil); err != nil {
					logging.Errorf("[TELEGRAF] failed to collect metric %s for host %d: %v", k, hid, err)
				}
			}
		}
	}

	switch v := raw.(type) {
	case []interface{}:
		for _, it := range v {
			if obj, ok := it.(map[string]interface{}); ok {
				processPoint(obj)
			}
		}
	case map[string]interface{}:
		processPoint(v)
	default:
		// unsupported
		return c.Status(400).JSON(fiber.Map{"error": "unsupported payload"})
	}

	return c.Status(201).JSON(fiber.Map{"msg": "ingested"})
}
