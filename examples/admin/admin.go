package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type AdminServer struct {
	store      *FlagStore
	sync       *SyncService
	yamlPath   string
	serverPort string
}

func NewAdminServer(store *FlagStore, yamlPath, port string) *AdminServer {
	return &AdminServer{
		store:      store,
		sync:       NewSyncService(store),
		yamlPath:   yamlPath,
		serverPort: port,
	}
}

func (s *AdminServer) Start() error {
	// Initial sync
	if err := s.sync.SyncToYAML(s.yamlPath); err != nil {
		return fmt.Errorf("initial sync: %w", err)
	}

	http.HandleFunc("/", s.handleUI)
	http.HandleFunc("/flags", s.handleFlags)
	http.HandleFunc("/flags/", s.handleFlag)

	log.Printf("Admin server starting on :%s", s.serverPort)
	log.Printf("Admin UI: http://localhost:%s", s.serverPort)
	log.Printf("YAML file: %s", s.yamlPath)
	return http.ListenAndServe(":"+s.serverPort, nil)
}

func (s *AdminServer) handleFlags(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listFlags(w, r)
	case http.MethodPost:
		s.createFlag(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *AdminServer) handleFlag(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[len("/flags/"):]
	if key == "" {
		http.Error(w, "Flag key required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getFlag(w, r, key)
	case http.MethodPut:
		s.updateFlag(w, r, key)
	case http.MethodDelete:
		s.deleteFlag(w, r, key)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *AdminServer) listFlags(w http.ResponseWriter, r *http.Request) {
	flags, err := s.store.GetAllFlags()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list flags: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flags)
}

func (s *AdminServer) getFlag(w http.ResponseWriter, r *http.Request, key string) {
	flag, err := s.store.GetFlag(key)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get flag: %v", err), http.StatusInternalServerError)
		return
	}

	if flag == nil {
		http.Error(w, "Flag not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flag)
}

func (s *AdminServer) createFlag(w http.ResponseWriter, r *http.Request) {
	var flag Flag
	if err := json.NewDecoder(r.Body).Decode(&flag); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Validate JSON fields
	if err := ValidateJSON(flag.Variants); err != nil {
		http.Error(w, fmt.Sprintf("Invalid variants JSON: %v", err), http.StatusBadRequest)
		return
	}
	if err := ValidateJSON(flag.Rules); err != nil {
		http.Error(w, fmt.Sprintf("Invalid rules JSON: %v", err), http.StatusBadRequest)
		return
	}
	if err := ValidateJSON(flag.Default); err != nil {
		http.Error(w, fmt.Sprintf("Invalid default JSON: %v", err), http.StatusBadRequest)
		return
	}

	if err := s.store.CreateFlag(flag); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create flag: %v", err), http.StatusInternalServerError)
		return
	}

	if err := s.sync.SyncToYAML(s.yamlPath); err != nil {
		log.Printf("Warning: failed to sync to YAML: %v", err)
		// Continue anyway
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(flag)
}

func (s *AdminServer) updateFlag(w http.ResponseWriter, r *http.Request, key string) {
	var flag Flag
	if err := json.NewDecoder(r.Body).Decode(&flag); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	flag.Key = key // Ensure key matches

	// Validate JSON fields
	if err := ValidateJSON(flag.Variants); err != nil {
		http.Error(w, fmt.Sprintf("Invalid variants JSON: %v", err), http.StatusBadRequest)
		return
	}
	if err := ValidateJSON(flag.Rules); err != nil {
		http.Error(w, fmt.Sprintf("Invalid rules JSON: %v", err), http.StatusBadRequest)
		return
	}
	if err := ValidateJSON(flag.Default); err != nil {
		http.Error(w, fmt.Sprintf("Invalid default JSON: %v", err), http.StatusBadRequest)
		return
	}

	if err := s.store.UpdateFlag(key, flag); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update flag: %v", err), http.StatusInternalServerError)
		return
	}

	if err := s.sync.SyncToYAML(s.yamlPath); err != nil {
		log.Printf("Warning: failed to sync to YAML: %v", err)
		// Continue anyway
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flag)
}

func (s *AdminServer) deleteFlag(w http.ResponseWriter, r *http.Request, key string) {
	if err := s.store.DeleteFlag(key); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete flag: %v", err), http.StatusInternalServerError)
		return
	}

	if err := s.sync.SyncToYAML(s.yamlPath); err != nil {
		log.Printf("Warning: failed to sync to YAML: %v", err)
		// Continue anyway
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *AdminServer) handleUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(adminHTML))
}

const adminHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Goff Admin - Feature Flags</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: #f5f5f5;
            color: #333;
            line-height: 1.6;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        header {
            background: white;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 {
            color: #2563eb;
            margin-bottom: 10px;
        }
        .subtitle {
            color: #666;
            font-size: 14px;
        }
        .controls {
            background: white;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .btn {
            background: #2563eb;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
            transition: background 0.2s;
        }
        .btn:hover { background: #1d4ed8; }
        .btn-danger { background: #dc2626; }
        .btn-danger:hover { background: #b91c1c; }
        .btn-secondary { background: #6b7280; }
        .btn-secondary:hover { background: #4b5563; }
        .flags-grid {
            display: grid;
            gap: 20px;
        }
        .flag-card {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            border-left: 4px solid #2563eb;
        }
        .flag-card.disabled {
            border-left-color: #9ca3af;
            opacity: 0.7;
        }
        .flag-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
        }
        .flag-key {
            font-size: 18px;
            font-weight: 600;
            color: #111;
        }
        .flag-type {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: 500;
            background: #dbeafe;
            color: #1e40af;
        }
        .flag-type.string { background: #fce7f3; color: #9f1239; }
        .flag-status {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: 500;
            background: #dcfce7;
            color: #166534;
        }
        .flag-status.disabled { background: #fee2e2; color: #991b1b; }
        .flag-details {
            margin-top: 15px;
            font-size: 14px;
            color: #666;
        }
        .flag-details pre {
            background: #f9fafb;
            padding: 10px;
            border-radius: 4px;
            overflow-x: auto;
            font-size: 12px;
            margin-top: 8px;
        }
        .flag-actions {
            margin-top: 15px;
            display: flex;
            gap: 10px;
        }
        .modal {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0,0,0,0.5);
            z-index: 1000;
            align-items: center;
            justify-content: center;
        }
        .modal.active { display: flex; }
        .modal-content {
            background: white;
            padding: 30px;
            border-radius: 8px;
            max-width: 600px;
            width: 90%;
            max-height: 90vh;
            overflow-y: auto;
        }
        .form-group {
            margin-bottom: 20px;
        }
        .form-group label {
            display: block;
            margin-bottom: 5px;
            font-weight: 500;
            color: #374151;
        }
        .form-group label:has(input[type="checkbox"]) {
            display: flex;
            align-items: center;
            gap: 8px;
            margin-bottom: 0;
        }
        .form-group label:has(input[type="checkbox"]) input[type="checkbox"] {
            width: auto;
            margin: 0;
        }
        .form-group input,
        .form-group select,
        .form-group textarea {
            width: 100%;
            padding: 10px;
            border: 1px solid #d1d5db;
            border-radius: 6px;
            font-size: 14px;
            font-family: monospace;
        }
        .form-group textarea {
            min-height: 100px;
            resize: vertical;
        }
        .form-actions {
            display: flex;
            gap: 10px;
            justify-content: flex-end;
            margin-top: 20px;
        }
        .error {
            background: #fee2e2;
            color: #991b1b;
            padding: 12px;
            border-radius: 6px;
            margin-bottom: 15px;
        }
        .success {
            background: #dcfce7;
            color: #166534;
            padding: 12px;
            border-radius: 6px;
            margin-bottom: 15px;
        }
        .loading {
            text-align: center;
            padding: 40px;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>🚩 Goff Feature Flags</h1>
            <div class="subtitle">Manage your feature flags</div>
        </header>

        <div class="controls">
            <button class="btn" onclick="loadFlags()">Refresh</button>
            <button class="btn" onclick="showCreateModal()">Create Flag</button>
        </div>

        <div id="message"></div>
        <div id="flags" class="flags-grid">
            <div class="loading">Loading flags...</div>
        </div>
    </div>

    <div id="modal" class="modal">
        <div class="modal-content">
            <h2 id="modal-title">Create Flag</h2>
            <div id="modal-message"></div>
            <form id="flag-form" onsubmit="saveFlag(event)">
                <div class="form-group">
                    <label>Key *</label>
                    <input type="text" id="flag-key" required pattern="[a-z0-9_]+" placeholder="new_feature">
                </div>
                <div class="form-group">
                    <label>Type *</label>
                    <select id="flag-type" required onchange="updateFormForType()">
                        <option value="bool">Boolean</option>
                        <option value="string">String</option>
                    </select>
                </div>
                <div class="form-group">
                    <label class="checkbox-label">
                        <input type="checkbox" id="flag-enabled" checked> Enabled
                    </label>
                </div>
                <div class="form-group">
                    <label>Default Value *</label>
                    <input type="text" id="flag-default" required placeholder='false or "default"'>
                </div>
                <div class="form-group">
                    <label>Variants (JSON) *</label>
                    <textarea id="flag-variants" required placeholder='{"true": 50, "false": 50}'></textarea>
                </div>
                <div class="form-group">
                    <label>Rules (JSON array)</label>
                    <textarea id="flag-rules" placeholder='[{"when": {"all": [{"attr": "plan", "op": "eq", "value": "pro"}]}, "then": {"variants": {"true": 90, "false": 10}}}]'></textarea>
                </div>
                <div class="form-actions">
                    <button type="button" class="btn btn-secondary" onclick="closeModal()">Cancel</button>
                    <button type="submit" class="btn">Save</button>
                </div>
            </form>
        </div>
    </div>

    <script>
        let editingKey = null;

        function showMessage(text, type) {
            const msg = document.getElementById('message');
            msg.className = type;
            msg.textContent = text;
            msg.style.display = 'block';
            setTimeout(() => msg.style.display = 'none', 5000);
        }

        async function loadFlags() {
            try {
                const res = await fetch('/flags');
                if (!res.ok) throw new Error('Failed to load flags');
                const flags = await res.json();
                renderFlags(flags);
            } catch (err) {
                showMessage('Error loading flags: ' + err.message, 'error');
                document.getElementById('flags').innerHTML = '<div class="error">Failed to load flags</div>';
            }
        }

        function renderFlags(flags) {
            const container = document.getElementById('flags');
            if (flags.length === 0) {
                container.innerHTML = '<div class="loading">No flags found. Create one to get started!</div>';
                return;
            }

            container.innerHTML = flags.map(flag => {
                const disabledClass = flag.enabled ? '' : 'disabled';
                const statusClass = flag.enabled ? '' : 'disabled';
                const statusText = flag.enabled ? 'Enabled' : 'Disabled';
                const prettyVariants = prettyJson(flag.variants);
                const prettyRules = flag.rules && flag.rules !== '[]' ? prettyJson(flag.rules) : '';
                const rulesHtml = prettyRules
                    ? '<div style="margin-top: 8px;"><strong>Rules:</strong></div><pre>' + escapeHtml(prettyRules) + '</pre>'
                    : '';
                return '<div class="flag-card ' + disabledClass + '">' +
                    '<div class="flag-header"><div>' +
                    '<span class="flag-key">' + escapeHtml(flag.key) + '</span> ' +
                    '<span class="flag-type ' + flag.type + '">' + flag.type + '</span> ' +
                    '<span class="flag-status ' + statusClass + '">' + statusText + '</span>' +
                    '</div></div>' +
                    '<div class="flag-details">' +
                    '<div><strong>Default:</strong> ' + escapeHtml(flag.default) + '</div>' +
                    '<div style="margin-top: 8px;"><strong>Variants:</strong></div>' +
                    '<pre>' + escapeHtml(prettyVariants) + '</pre>' +
                    rulesHtml +
                    '</div>' +
                    '<div class="flag-actions">' +
                    '<button class="btn" onclick="editFlag(\'' + escapeHtml(flag.key) + '\')">Edit</button> ' +
                    '<button class="btn btn-danger" onclick="deleteFlag(\'' + escapeHtml(flag.key) + '\')">Delete</button>' +
                    '</div></div>';
            }).join('');
        }

        function showCreateModal() {
            editingKey = null;
            document.getElementById('modal-title').textContent = 'Create Flag';
            document.getElementById('flag-form').reset();
            document.getElementById('flag-enabled').checked = true;
            document.getElementById('flag-type').value = 'bool';
            updateFormForType();
            document.getElementById('modal').classList.add('active');
        }

        function editFlag(key) {
            editingKey = key;
            fetch('/flags/' + encodeURIComponent(key))
                .then(res => res.json())
                .then(flag => {
                    document.getElementById('modal-title').textContent = 'Edit Flag';
                    document.getElementById('flag-key').value = flag.key;
                    document.getElementById('flag-key').disabled = true;
                    document.getElementById('flag-type').value = flag.type;
                    document.getElementById('flag-enabled').checked = flag.enabled;
                    document.getElementById('flag-default').value = flag.default;
                    document.getElementById('flag-variants').value = flag.variants;
                    document.getElementById('flag-rules').value = flag.rules;
                    updateFormForType();
                    document.getElementById('modal').classList.add('active');
                })
                .catch(err => showMessage('Error loading flag: ' + err.message, 'error'));
        }

        function updateFormForType() {
            const type = document.getElementById('flag-type').value;
            const defaultInput = document.getElementById('flag-default');
            if (type === 'bool') {
                defaultInput.placeholder = 'true or false';
            } else {
                defaultInput.placeholder = '"default_value"';
            }
        }

        function closeModal() {
            document.getElementById('modal').classList.remove('active');
            document.getElementById('flag-key').disabled = false;
            document.getElementById('modal-message').innerHTML = '';
        }

        async function saveFlag(e) {
            e.preventDefault();
            const msgEl = document.getElementById('modal-message');
            msgEl.innerHTML = '';

            const flag = {
                key: document.getElementById('flag-key').value,
                type: document.getElementById('flag-type').value,
                enabled: document.getElementById('flag-enabled').checked,
                default: document.getElementById('flag-default').value,
                variants: document.getElementById('flag-variants').value,
                rules: document.getElementById('flag-rules').value || '[]'
            };

            try {
                // Validate JSON
                JSON.parse(flag.variants);
                JSON.parse(flag.rules);
                if (flag.type === 'bool') {
                    JSON.parse(flag.default);
                } else {
                    JSON.parse(flag.default);
                }
            } catch (err) {
                msgEl.innerHTML = '<div class="error">Invalid JSON: ' + err.message + '</div>';
                return;
            }

            const url = editingKey ? '/flags/' + encodeURIComponent(editingKey) : '/flags';
            const method = editingKey ? 'PUT' : 'POST';

            try {
                const res = await fetch(url, {
                    method,
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(flag)
                });

                if (!res.ok) {
                    const err = await res.text();
                    throw new Error(err);
                }

                showMessage(editingKey ? 'Flag updated successfully' : 'Flag created successfully', 'success');
                closeModal();
                loadFlags();
            } catch (err) {
                msgEl.innerHTML = '<div class="error">Error: ' + err.message + '</div>';
            }
        }

        async function deleteFlag(key) {
            if (!confirm('Are you sure you want to delete flag "' + key + '"?')) return;

            try {
                const res = await fetch('/flags/' + encodeURIComponent(key), { method: 'DELETE' });
                if (!res.ok) throw new Error('Failed to delete flag');
                showMessage('Flag deleted successfully', 'success');
                loadFlags();
            } catch (err) {
                showMessage('Error deleting flag: ' + err.message, 'error');
            }
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        function prettyJson(jsonString) {
            try {
                const parsed = JSON.parse(jsonString);
                return JSON.stringify(parsed, null, 2);
            } catch (err) {
                return jsonString;
            }
        }

        // Load flags on page load
        loadFlags();
    </script>
</body>
</html>`
