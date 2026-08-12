<template>
  <div class="min-h-screen bg-gray-950 text-gray-100">
    <!-- Header -->
    <div class="bg-gray-900 border-b border-gray-800 px-6 py-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-red-500 to-rose-700 flex items-center justify-center shadow-lg">
            <ShieldAlert :size="20" class="text-white" />
          </div>
          <div>
            <h1 class="text-xl font-bold text-white">System Diagnostics</h1>
            <p class="text-xs text-gray-400 mt-0.5">Error Tracking — SQL, API, Backend, Frontend, HTTP</p>
          </div>
        </div>
        <div class="flex items-center gap-3">
          <select v-model="statsPeriod" @change="loadStats" class="bg-gray-800 border border-gray-700 text-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-red-500">
            <option value="24h">Last 24h</option>
            <option value="7d">Last 7 days</option>
            <option value="30d">Last 30 days</option>
            <option value="90d">Last 90 days</option>
          </select>
          <button @click="openCreateModal" class="flex items-center gap-2 px-4 py-2 bg-gray-700 hover:bg-gray-600 text-gray-200 rounded-lg text-sm font-medium transition-colors">
            <Plus :size="15" /> Ajouter Log
          </button>
          <button @click="openPurgeModal" class="flex items-center gap-2 px-4 py-2 bg-red-900/50 hover:bg-red-800/60 text-red-300 rounded-lg text-sm font-medium transition-colors border border-red-700/40">
            <Trash2 :size="15" /> Purger
          </button>
        </div>
      </div>
    </div>

    <div class="p-6 space-y-6">

      <!-- KPI Banner -->
      <div class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-7 gap-3">
        <div class="bg-gray-900 border border-gray-800 rounded-xl p-3 text-center">
          <p class="text-xs text-gray-500 mb-1">Total</p>
          <p class="text-xl font-black text-white">{{ stats.total ?? 0 }}</p>
        </div>
        <div class="bg-red-950/40 border border-red-800/40 rounded-xl p-3 text-center">
          <p class="text-xs text-red-400 mb-1">Critical</p>
          <p class="text-xl font-black text-red-400">{{ stats.critical ?? 0 }}</p>
        </div>
        <div class="bg-rose-950/30 border border-rose-800/30 rounded-xl p-3 text-center">
          <p class="text-xs text-rose-400 mb-1">Errors</p>
          <p class="text-xl font-black text-rose-400">{{ stats.errors ?? 0 }}</p>
        </div>
        <div class="bg-yellow-950/30 border border-yellow-800/30 rounded-xl p-3 text-center">
          <p class="text-xs text-yellow-400 mb-1">Warnings</p>
          <p class="text-xl font-black text-yellow-400">{{ stats.warnings ?? 0 }}</p>
        </div>
        <div class="bg-blue-950/30 border border-blue-800/30 rounded-xl p-3 text-center">
          <p class="text-xs text-blue-400 mb-1">Info</p>
          <p class="text-xl font-black text-blue-400">{{ stats.info ?? 0 }}</p>
        </div>
        <div class="bg-orange-950/30 border border-orange-800/30 rounded-xl p-3 text-center">
          <p class="text-xs text-orange-400 mb-1">Unresolved</p>
          <p class="text-xl font-black text-orange-400">{{ stats.unresolved ?? 0 }}</p>
        </div>
        <div class="bg-emerald-950/30 border border-emerald-800/30 rounded-xl p-3 text-center">
          <p class="text-xs text-emerald-400 mb-1">Resolved Today</p>
          <p class="text-xl font-black text-emerald-400">{{ stats.resolved_today ?? 0 }}</p>
        </div>
      </div>

      <!-- Charts Row -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <!-- Hourly Trend -->
        <div class="lg:col-span-2 bg-gray-900 border border-gray-800 rounded-xl p-4">
          <div class="flex items-center gap-2 mb-4">
            <TrendingUp :size="15" class="text-blue-400" />
            <span class="text-sm font-semibold text-gray-200">Tendance Horaire (24h)</span>
          </div>
          <div v-if="(stats.hourly_trend || []).length > 0" class="overflow-x-auto">
            <div class="flex items-end gap-1 h-28 min-w-[400px]">
              <div v-for="h in stats.hourly_trend" :key="h.hour" class="flex-1 flex flex-col items-center gap-0.5">
                <div class="w-full flex flex-col items-center relative group">
                  <div class="w-full bg-blue-600/50 hover:bg-blue-500 rounded-t transition-all"
                    :style="{ height: maxHourly > 0 ? `${(h.count / maxHourly) * 80}px` : '2px', minHeight: h.count > 0 ? '3px' : '0' }"></div>
                  <div class="absolute bottom-full mb-1 hidden group-hover:block bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs text-gray-200 whitespace-nowrap z-20">
                    {{ fmtHour(h.hour) }}: {{ h.count }} logs ({{ h.errors }} errors)
                  </div>
                </div>
                <span class="text-gray-600 text-xs">{{ fmtHourShort(h.hour) }}</span>
              </div>
            </div>
          </div>
          <div v-else class="h-28 flex items-center justify-center text-gray-600 text-sm">Aucune donnée</div>
        </div>

        <!-- By Source -->
        <div class="bg-gray-900 border border-gray-800 rounded-xl p-4">
          <div class="flex items-center gap-2 mb-4">
            <Layers :size="15" class="text-purple-400" />
            <span class="text-sm font-semibold text-gray-200">Par Source</span>
          </div>
          <div class="space-y-2">
            <div v-for="s in (stats.by_source || []).slice(0,8)" :key="s.source" class="flex items-center gap-2">
              <span :class="['w-2 h-2 rounded-full flex-shrink-0', sourceDot(s.source)]"></span>
              <div class="flex-1 min-w-0">
                <div class="flex justify-between mb-0.5">
                  <span class="text-xs text-gray-300 truncate">{{ sourceLabel(s.source) }}</span>
                  <span class="text-xs font-bold text-white ml-2">{{ s.count }}</span>
                </div>
                <div class="h-1 bg-gray-800 rounded-full overflow-hidden">
                  <div :class="['h-full rounded-full', sourceBar(s.source)]"
                    :style="{ width: stats.total > 0 ? `${(s.count / stats.total) * 100}%` : '0%' }"></div>
                </div>
              </div>
            </div>
            <div v-if="!(stats.by_source || []).length" class="text-center text-gray-600 text-sm py-4">Aucune donnée</div>
          </div>
        </div>
      </div>

      <!-- Top Errors + By Module -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <!-- Top Recurring Errors -->
        <div class="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
          <div class="px-4 py-3 border-b border-gray-800 flex items-center gap-2">
            <AlertOctagon :size="15" class="text-red-400" />
            <span class="text-sm font-semibold text-gray-200">Top 10 Erreurs Récurrentes</span>
          </div>
          <div class="divide-y divide-gray-800">
            <div v-for="(e, i) in (stats.top_errors || []).slice(0,10)" :key="i"
              class="px-4 py-3 hover:bg-gray-800/50 transition-colors cursor-pointer"
              @click="searchByMessage(e.message)">
              <div class="flex items-start justify-between gap-2">
                <div class="flex-1 min-w-0">
                  <p class="text-xs text-gray-200 font-medium truncate">{{ e.message }}</p>
                  <div class="flex items-center gap-2 mt-1">
                    <span v-if="e.module" class="px-1.5 py-0.5 bg-gray-800 text-gray-400 rounded text-xs">{{ e.module }}</span>
                    <span v-if="e.http_status > 0" :class="['px-1.5 py-0.5 rounded text-xs font-mono font-bold', httpStatusClass(e.http_status)]">{{ e.http_status }}</span>
                  </div>
                </div>
                <div class="text-right flex-shrink-0">
                  <span class="px-2 py-0.5 bg-red-900/50 text-red-300 rounded-full text-xs font-bold">{{ e.count }}x</span>
                  <p class="text-xs text-gray-600 mt-1">{{ fmtRelative(e.last_seen) }}</p>
                </div>
              </div>
            </div>
            <div v-if="!(stats.top_errors || []).length" class="px-4 py-8 text-center text-gray-600 text-sm">Aucune erreur</div>
          </div>
        </div>

        <!-- By Module -->
        <div class="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
          <div class="px-4 py-3 border-b border-gray-800 flex items-center gap-2">
            <Box :size="15" class="text-cyan-400" />
            <span class="text-sm font-semibold text-gray-200">Par Module</span>
          </div>
          <div class="divide-y divide-gray-800">
            <div v-for="m in (stats.by_module || []).slice(0,10)" :key="m.module" class="px-4 py-3">
              <div class="flex items-center justify-between mb-1">
                <span class="text-sm text-gray-200 font-medium">{{ m.module }}</span>
                <div class="flex items-center gap-2">
                  <span class="text-xs font-bold text-white">{{ m.count }}</span>
                  <span :class="['text-xs font-bold', m.error_rate > 50 ? 'text-red-400' : m.error_rate > 20 ? 'text-yellow-400' : 'text-emerald-400']">
                    {{ m.error_rate }}%
                  </span>
                </div>
              </div>
              <div class="h-1.5 bg-gray-800 rounded-full overflow-hidden">
                <div class="h-full bg-cyan-600 rounded-full" :style="{ width: stats.total > 0 ? `${(m.count / stats.total) * 100}%` : '0%' }"></div>
              </div>
            </div>
            <div v-if="!(stats.by_module || []).length" class="px-4 py-8 text-center text-gray-600 text-sm">Aucune donnée</div>
          </div>
        </div>
      </div>

      <!-- Filters -->
      <div class="bg-gray-900 border border-gray-800 rounded-xl p-4">
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex items-center gap-2">
            <Filter :size="15" class="text-gray-400" />
            <span class="text-xs text-gray-400 font-medium">Filtres :</span>
          </div>
          <div class="flex gap-1">
            <button v-for="sev in severityFilters" :key="sev.value"
              @click="filterSeverity = sev.value; loadLogs()"
              :class="['px-3 py-1.5 rounded-lg text-xs font-medium transition-colors',
                filterSeverity === sev.value ? sev.active : 'bg-gray-800 text-gray-400 hover:text-gray-200']">
              {{ sev.label }}
            </button>
          </div>
          <select v-model="filterSource" @change="loadLogs()" class="bg-gray-800 border border-gray-700 text-gray-200 rounded-lg px-3 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-red-500">
            <option value="">Toutes sources</option>
            <option value="frontend_js">Frontend JS</option>
            <option value="backend_go">Backend Go</option>
            <option value="database_sql">Database SQL</option>
            <option value="api_http">API HTTP</option>
            <option value="auth">Auth</option>
            <option value="system">System</option>
          </select>
          <select v-model="filterResolved" @change="loadLogs()" class="bg-gray-800 border border-gray-700 text-gray-200 rounded-lg px-3 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-red-500">
            <option value="">Tous statuts</option>
            <option value="false">Non résolus</option>
            <option value="true">Résolus</option>
          </select>
          <input v-model="filterModule" @input="debouncedLoad" type="text" placeholder="Module..." class="bg-gray-800 border border-gray-700 text-gray-200 rounded-lg px-3 py-1.5 text-xs w-32 focus:outline-none focus:ring-1 focus:ring-red-500 placeholder-gray-600" />
          <div class="relative flex-1 min-w-[200px]">
            <Search :size="13" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500" />
            <input v-model="filterSearch" @input="debouncedLoad" type="text" placeholder="Rechercher message, endpoint..." class="w-full bg-gray-800 border border-gray-700 text-gray-200 rounded-lg pl-8 pr-3 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-red-500 placeholder-gray-600" />
          </div>
          <div class="flex items-center gap-2 ml-auto">
            <template v-if="selectedIds.length > 0">
              <span class="text-xs text-gray-400">{{ selectedIds.length }} sélectionné(s)</span>
              <button @click="bulkResolve" class="px-3 py-1.5 bg-emerald-700 hover:bg-emerald-600 text-white rounded-lg text-xs font-medium transition-colors">
                Résoudre
              </button>
            </template>
            <button @click="loadLogs" class="p-1.5 bg-gray-800 rounded-lg hover:bg-gray-700 text-gray-400 transition-colors">
              <RefreshCw :size="14" />
            </button>
          </div>
        </div>
      </div>

      <!-- Logs Table -->
      <div v-if="loading" class="flex items-center justify-center py-16">
        <div class="flex flex-col items-center gap-3">
          <div class="w-10 h-10 border-2 border-red-500 border-t-transparent rounded-full animate-spin"></div>
          <span class="text-sm text-gray-400">Chargement des logs...</span>
        </div>
      </div>

      <div v-else class="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
        <div class="px-5 py-3 border-b border-gray-800 flex items-center justify-between">
          <div class="flex items-center gap-2">
            <ListX :size="15" class="text-red-400" />
            <span class="text-sm font-semibold text-gray-200">Logs Système</span>
            <span class="px-2 py-0.5 bg-gray-800 rounded-full text-xs text-gray-400">{{ logsData.total }} entrées</span>
          </div>
          <div class="flex items-center gap-2">
            <input type="checkbox" :checked="allSelected" @change="toggleSelectAll" class="w-3.5 h-3.5 accent-red-500" />
            <span class="text-xs text-gray-500">Tout sélectionner</span>
          </div>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-gray-800/60">
                <th class="px-3 py-3 w-8"></th>
                <th class="px-3 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wide">Sévérité</th>
                <th class="px-3 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wide">Source</th>
                <th class="px-3 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wide">Module</th>
                <th class="px-3 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wide w-80">Message</th>
                <th class="px-3 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wide">Endpoint</th>
                <th class="px-3 py-3 text-center text-xs font-semibold text-gray-400 uppercase tracking-wide">HTTP</th>
                <th class="px-3 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wide">IP</th>
                <th class="px-3 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wide">Date</th>
                <th class="px-3 py-3 text-center text-xs font-semibold text-gray-400 uppercase tracking-wide">Statut</th>
                <th class="px-3 py-3 text-center text-xs font-semibold text-gray-400 uppercase tracking-wide">Actions</th>
              </tr>
            </thead>
            <tbody>
              <template v-if="logsData.logs.length > 0">
                <tr v-for="log in logsData.logs" :key="log.id"
                  :class="['border-t border-gray-800 hover:bg-gray-800/40 transition-colors cursor-pointer',
                    log.severity === 'critical' ? 'bg-red-950/10' : '']"
                  @click.stop="viewLog(log)">
                  <td class="px-3 py-3" @click.stop>
                    <input type="checkbox" :value="log.id" v-model="selectedIds" class="w-3.5 h-3.5 accent-red-500" />
                  </td>
                  <td class="px-3 py-3">
                    <span :class="['px-2 py-0.5 rounded-full text-xs font-bold', severityBadge(log.severity)]">
                      {{ log.severity }}
                    </span>
                  </td>
                  <td class="px-3 py-3">
                    <span :class="['px-2 py-0.5 rounded text-xs', sourceBadge(log.source)]">
                      {{ sourceLabel(log.source) }}
                    </span>
                  </td>
                  <td class="px-3 py-3 text-gray-400 text-xs">{{ log.module || '—' }}</td>
                  <td class="px-3 py-3">
                    <p class="text-gray-200 text-xs font-medium line-clamp-2 max-w-[300px]">{{ log.message }}</p>
                    <p v-if="log.error_code" class="text-gray-500 text-xs font-mono mt-0.5">{{ log.error_code }}</p>
                  </td>
                  <td class="px-3 py-3 text-gray-500 text-xs font-mono truncate max-w-[150px]">{{ log.endpoint || '—' }}</td>
                  <td class="px-3 py-3 text-center">
                    <span v-if="log.http_status > 0" :class="['px-1.5 py-0.5 rounded text-xs font-mono font-bold', httpStatusClass(log.http_status)]">
                      {{ log.http_status }}
                    </span>
                    <span v-else class="text-gray-600">—</span>
                  </td>
                  <td class="px-3 py-3 text-gray-500 text-xs font-mono">{{ log.ip_address || '—' }}</td>
                  <td class="px-3 py-3 text-gray-400 text-xs whitespace-nowrap">{{ fmtRelative(log.created_at) }}</td>
                  <td class="px-3 py-3 text-center">
                    <span v-if="log.is_resolved" class="px-2 py-0.5 bg-emerald-900/50 text-emerald-300 rounded-full text-xs">Résolu</span>
                    <span v-else class="px-2 py-0.5 bg-orange-900/50 text-orange-300 rounded-full text-xs">Ouvert</span>
                  </td>
                  <td class="px-3 py-3 text-center" @click.stop>
                    <div class="flex items-center justify-center gap-1">
                      <button @click="viewLog(log)" class="p-1.5 rounded-lg hover:bg-gray-700 text-gray-400 hover:text-blue-400 transition-colors">
                        <Eye :size="13" />
                      </button>
                      <button v-if="!log.is_resolved" @click="resolveLog(log.id)" class="p-1.5 rounded-lg hover:bg-gray-700 text-gray-400 hover:text-emerald-400 transition-colors">
                        <CheckCircle2 :size="13" />
                      </button>
                      <button @click="deleteLog(log.id)" class="p-1.5 rounded-lg hover:bg-gray-700 text-gray-400 hover:text-red-400 transition-colors">
                        <Trash2 :size="13" />
                      </button>
                    </div>
                  </td>
                </tr>
              </template>
              <tr v-else>
                <td colspan="11" class="px-4 py-16 text-center">
                  <div class="flex flex-col items-center gap-3">
                    <div class="w-14 h-14 rounded-full bg-gray-800 flex items-center justify-center">
                      <ShieldCheck :size="26" class="text-emerald-400" />
                    </div>
                    <p class="text-gray-400 font-medium">Aucun log pour ces filtres</p>
                    <p class="text-gray-600 text-sm">Le système fonctionne normalement</p>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <!-- Pagination -->
        <div v-if="logsData.pages > 1" class="px-5 py-3 border-t border-gray-800 flex items-center justify-between">
          <span class="text-xs text-gray-500">Page {{ logsData.page }} / {{ logsData.pages }} — {{ logsData.total }} entrées</span>
          <div class="flex gap-2">
            <button @click="currentPage--; loadLogs()" :disabled="currentPage <= 1" class="px-3 py-1.5 rounded-lg bg-gray-800 border border-gray-700 text-sm text-gray-300 disabled:opacity-40 hover:bg-gray-700">
              <ChevronLeft :size="14" />
            </button>
            <button @click="currentPage++; loadLogs()" :disabled="currentPage >= logsData.pages" class="px-3 py-1.5 rounded-lg bg-gray-800 border border-gray-700 text-sm text-gray-300 disabled:opacity-40 hover:bg-gray-700">
              <ChevronRight :size="14" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Detail Modal -->
    <Teleport to="body">
      <div v-if="showDetailModal && selectedLog" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm">
        <div class="bg-gray-900 border border-gray-700 rounded-2xl shadow-2xl w-full max-w-3xl max-h-[90vh] overflow-y-auto">
          <div class="flex items-center justify-between p-5 border-b border-gray-800">
            <div class="flex items-center gap-3">
              <div :class="['w-8 h-8 rounded-lg flex items-center justify-center', severityBg(selectedLog.severity)]">
                <component :is="severityIcon(selectedLog.severity)" :size="16" class="text-white" />
              </div>
              <div>
                <h2 class="text-base font-bold text-white">Détail du Log</h2>
                <p class="text-xs text-gray-500 font-mono">{{ selectedLog.id }}</p>
              </div>
            </div>
            <button @click="showDetailModal = false" class="p-2 rounded-lg hover:bg-gray-800 text-gray-400">
              <X :size="18" />
            </button>
          </div>
          <div class="p-5 space-y-4">
            <!-- Badges row -->
            <div class="flex flex-wrap gap-2">
              <span :class="['px-2.5 py-1 rounded-full text-xs font-bold', severityBadge(selectedLog.severity)]">{{ selectedLog.severity }}</span>
              <span :class="['px-2.5 py-1 rounded text-xs', sourceBadge(selectedLog.source)]">{{ sourceLabel(selectedLog.source) }}</span>
              <span v-if="selectedLog.module" class="px-2.5 py-1 bg-gray-800 text-gray-300 rounded text-xs">{{ selectedLog.module }}</span>
              <span v-if="selectedLog.http_status > 0" :class="['px-2.5 py-1 rounded text-xs font-mono font-bold', httpStatusClass(selectedLog.http_status)]">HTTP {{ selectedLog.http_status }}</span>
              <span :class="['px-2.5 py-1 rounded-full text-xs', selectedLog.is_resolved ? 'bg-emerald-900/50 text-emerald-300' : 'bg-orange-900/50 text-orange-300']">
                {{ selectedLog.is_resolved ? 'Résolu' : 'Non résolu' }}
              </span>
            </div>

            <!-- Message -->
            <div class="bg-gray-800 rounded-xl p-4">
              <p class="text-xs text-gray-500 mb-1 font-semibold uppercase tracking-wide">Message</p>
              <p class="text-gray-100 font-medium">{{ selectedLog.message }}</p>
              <p v-if="selectedLog.error_code" class="text-red-400 font-mono text-xs mt-1">{{ selectedLog.error_code }}</p>
              <p v-if="selectedLog.sql_state" class="text-amber-400 font-mono text-xs mt-1">SQLSTATE: {{ selectedLog.sql_state }}</p>
            </div>

            <!-- Meta grid -->
            <div class="grid grid-cols-2 md:grid-cols-3 gap-3 text-sm">
              <div class="bg-gray-800 rounded-xl p-3">
                <p class="text-xs text-gray-500 mb-1">Date / Heure</p>
                <p class="text-gray-200 font-medium">{{ fmtDateTime(selectedLog.created_at) }}</p>
              </div>
              <div class="bg-gray-800 rounded-xl p-3">
                <p class="text-xs text-gray-500 mb-1">IP Address</p>
                <p class="text-gray-200 font-mono text-xs">{{ selectedLog.ip_address || '—' }}</p>
              </div>
              <div class="bg-gray-800 rounded-xl p-3">
                <p class="text-xs text-gray-500 mb-1">Duration</p>
                <p class="text-gray-200 font-medium">{{ selectedLog.duration_ms > 0 ? selectedLog.duration_ms + 'ms' : '—' }}</p>
              </div>
              <div class="bg-gray-800 rounded-xl p-3 col-span-2">
                <p class="text-xs text-gray-500 mb-1">Endpoint</p>
                <p class="text-gray-200 font-mono text-xs">{{ selectedLog.endpoint || '—' }}</p>
              </div>
              <div class="bg-gray-800 rounded-xl p-3">
                <p class="text-xs text-gray-500 mb-1">Request ID</p>
                <p class="text-gray-400 font-mono text-xs">{{ selectedLog.request_id || '—' }}</p>
              </div>
            </div>

            <!-- Page URL -->
            <div v-if="selectedLog.page_url" class="bg-gray-800 rounded-xl p-3">
              <p class="text-xs text-gray-500 mb-1 font-semibold uppercase tracking-wide">Page URL</p>
              <p class="text-blue-400 font-mono text-xs">{{ selectedLog.page_url }}</p>
            </div>

            <!-- Stack Trace -->
            <div v-if="selectedLog.stack_trace" class="bg-gray-950 border border-gray-700 rounded-xl p-4">
              <p class="text-xs text-gray-500 mb-2 font-semibold uppercase tracking-wide flex items-center gap-1">
                <Terminal :size="12" /> Stack Trace
              </p>
              <pre class="text-xs text-gray-300 whitespace-pre-wrap font-mono overflow-x-auto max-h-60 overflow-y-auto">{{ selectedLog.stack_trace }}</pre>
            </div>

            <!-- User Agent -->
            <div v-if="selectedLog.user_agent" class="bg-gray-800 rounded-xl p-3">
              <p class="text-xs text-gray-500 mb-1 font-semibold uppercase tracking-wide">User Agent</p>
              <p class="text-gray-400 text-xs">{{ selectedLog.user_agent }}</p>
            </div>

            <!-- Resolution -->
            <div v-if="selectedLog.is_resolved && selectedLog.resolution_note" class="bg-emerald-950/30 border border-emerald-800/30 rounded-xl p-3">
              <p class="text-xs text-emerald-400 font-semibold mb-1">Note de résolution</p>
              <p class="text-emerald-300 text-sm">{{ selectedLog.resolution_note }}</p>
            </div>

            <!-- Actions -->
            <div class="flex gap-3 pt-2">
              <button @click="showDetailModal = false" class="flex-1 px-4 py-2.5 bg-gray-800 border border-gray-700 text-gray-200 rounded-xl text-sm font-medium hover:bg-gray-700 transition-colors">
                Fermer
              </button>
              <button v-if="!selectedLog.is_resolved" @click="resolveLog(selectedLog.id); showDetailModal = false" class="flex-1 px-4 py-2.5 bg-emerald-700 hover:bg-emerald-600 text-white rounded-xl text-sm font-medium transition-colors flex items-center justify-center gap-2">
                <CheckCircle2 :size="15" /> Marquer résolu
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Create Log Modal -->
    <Teleport to="body">
      <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm">
        <div class="bg-gray-900 border border-gray-700 rounded-2xl shadow-2xl w-full max-w-xl max-h-[90vh] overflow-y-auto">
          <div class="flex items-center justify-between p-5 border-b border-gray-800">
            <h2 class="text-lg font-bold text-white">Ajouter Log Manuel</h2>
            <button @click="showCreateModal = false" class="p-2 rounded-lg hover:bg-gray-800 text-gray-400"><X :size="18" /></button>
          </div>
          <form @submit.prevent="submitCreateLog" class="p-5 space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="lbl">Sévérité *</label>
                <select v-model="createForm.severity" class="field">
                  <option value="debug">Debug</option>
                  <option value="info">Info</option>
                  <option value="warning">Warning</option>
                  <option value="error">Error</option>
                  <option value="critical">Critical</option>
                </select>
              </div>
              <div>
                <label class="lbl">Source *</label>
                <select v-model="createForm.source" class="field">
                  <option value="frontend_js">Frontend JS</option>
                  <option value="backend_go">Backend Go</option>
                  <option value="database_sql">Database SQL</option>
                  <option value="api_http">API HTTP</option>
                  <option value="auth">Auth</option>
                  <option value="system">System</option>
                </select>
              </div>
              <div>
                <label class="lbl">Module</label>
                <input v-model="createForm.module" type="text" placeholder="treasury, tax, sales..." class="field" />
              </div>
              <div>
                <label class="lbl">HTTP Status</label>
                <input v-model.number="createForm.http_status" type="number" placeholder="500, 404, 403..." class="field" />
              </div>
              <div>
                <label class="lbl">Endpoint</label>
                <input v-model="createForm.endpoint" type="text" placeholder="/api/v1/..." class="field" />
              </div>
              <div>
                <label class="lbl">Error Code / SQLSTATE</label>
                <input v-model="createForm.error_code" type="text" placeholder="42703, ERR_001..." class="field" />
              </div>
              <div class="col-span-2">
                <label class="lbl">Message *</label>
                <textarea v-model="createForm.message" rows="3" required class="field resize-none" placeholder="Description de l'erreur..."></textarea>
              </div>
              <div class="col-span-2">
                <label class="lbl">Stack Trace</label>
                <textarea v-model="createForm.stack_trace" rows="4" class="field resize-none font-mono text-xs" placeholder="Stack trace complet..."></textarea>
              </div>
            </div>
            <div class="flex gap-3">
              <button type="button" @click="showCreateModal = false" class="flex-1 px-4 py-2.5 bg-gray-800 border border-gray-700 text-gray-200 rounded-xl text-sm font-medium hover:bg-gray-700 transition-colors">
                Annuler
              </button>
              <button type="submit" :disabled="submitting" class="flex-1 px-4 py-2.5 bg-red-700 hover:bg-red-600 text-white rounded-xl text-sm font-bold transition-colors disabled:opacity-60 flex items-center justify-center gap-2">
                <div v-if="submitting" class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                <Save v-else :size="15" /> Enregistrer
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- Purge Modal -->
    <Teleport to="body">
      <div v-if="showPurgeModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm">
        <div class="bg-gray-900 border border-red-700/40 rounded-2xl shadow-2xl w-full max-w-md">
          <div class="flex items-center justify-between p-5 border-b border-gray-800">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-lg bg-red-700 flex items-center justify-center">
                <Trash2 :size="16" class="text-white" />
              </div>
              <h2 class="text-lg font-bold text-white">Purger les Logs</h2>
            </div>
            <button @click="showPurgeModal = false" class="p-2 rounded-lg hover:bg-gray-800 text-gray-400"><X :size="18" /></button>
          </div>
          <div class="p-5 space-y-4">
            <div>
              <label class="lbl">Supprimer les logs de plus de (jours)</label>
              <input v-model.number="purgeForm.older_than_days" type="number" min="1" class="field" />
            </div>
            <div>
              <label class="lbl">Sévérité (optionnel)</label>
              <select v-model="purgeForm.severity" class="field">
                <option value="">Toutes</option>
                <option value="debug">Debug</option>
                <option value="info">Info</option>
                <option value="warning">Warning</option>
              </select>
            </div>
            <label class="flex items-center gap-3 cursor-pointer">
              <input type="checkbox" v-model="purgeForm.only_resolved" class="w-4 h-4 accent-red-500" />
              <span class="text-sm text-gray-300">Uniquement les logs résolus</span>
            </label>
            <div class="bg-red-950/30 border border-red-800/30 rounded-xl p-3">
              <p class="text-xs text-red-300 font-medium">Attention: Cette action est irréversible.</p>
            </div>
            <div class="flex gap-3">
              <button @click="showPurgeModal = false" class="flex-1 px-4 py-2.5 bg-gray-800 border border-gray-700 text-gray-200 rounded-xl text-sm font-medium hover:bg-gray-700 transition-colors">
                Annuler
              </button>
              <button @click="submitPurge" :disabled="submitting" class="flex-1 px-4 py-2.5 bg-red-700 hover:bg-red-600 text-white rounded-xl text-sm font-bold transition-colors disabled:opacity-60 flex items-center justify-center gap-2">
                <Trash2 :size="15" /> Purger
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import {
  ShieldAlert, ShieldCheck, Plus, Trash2, Filter, Search, RefreshCw,
  Eye, CheckCircle2, X, Save, TrendingUp, Layers, AlertOctagon, Box,
  ListX, ChevronLeft, ChevronRight, Terminal, AlertCircle, AlertTriangle,
  Info, Bug
} from 'lucide-vue-next'
import { diagnosticsAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const store = useAppStore()
const loading = ref(false)
const submitting = ref(false)
const statsPeriod = ref('7d')

const stats = ref<any>({})
const logsData = ref<any>({ logs: [], total: 0, page: 1, pages: 1 })
const currentPage = ref(1)

const filterSeverity = ref('')
const filterSource = ref('')
const filterResolved = ref('')
const filterModule = ref('')
const filterSearch = ref('')
const selectedIds = ref<string[]>([])

const showDetailModal = ref(false)
const showCreateModal = ref(false)
const showPurgeModal = ref(false)
const selectedLog = ref<any>(null)

const severityFilters = [
  { value: '', label: 'Tous', active: 'bg-gray-600 text-white' },
  { value: 'critical', label: 'Critical', active: 'bg-red-700 text-white' },
  { value: 'error', label: 'Error', active: 'bg-rose-700 text-white' },
  { value: 'warning', label: 'Warning', active: 'bg-yellow-700 text-yellow-100' },
  { value: 'info', label: 'Info', active: 'bg-blue-700 text-white' },
  { value: 'debug', label: 'Debug', active: 'bg-gray-700 text-white' },
]

const defaultCreateForm = () => ({
  severity: 'error',
  source: 'backend_go',
  module: '',
  endpoint: '',
  method: '',
  http_status: 0,
  message: '',
  error_code: '',
  stack_trace: ''
})
const createForm = ref(defaultCreateForm())
const purgeForm = ref({ older_than_days: 90, severity: '', only_resolved: false })

let debounceTimer: ReturnType<typeof setTimeout>
function debouncedLoad() {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => { currentPage.value = 1; loadLogs() }, 400)
}

function searchByMessage(msg: string) {
  filterSearch.value = msg
  currentPage.value = 1
  loadLogs()
}

async function loadStats() {
  try {
    const res = await diagnosticsAPI.getStats(statsPeriod.value)
    stats.value = res.data || {}
  } catch { stats.value = {} }
}

async function loadLogs() {
  loading.value = true
  selectedIds.value = []
  try {
    const params: any = { page: currentPage.value, limit: 50 }
    if (filterSeverity.value) params.severity = filterSeverity.value
    if (filterSource.value) params.source = filterSource.value
    if (filterResolved.value) params.resolved = filterResolved.value
    if (filterModule.value) params.module = filterModule.value
    if (filterSearch.value) params.search = filterSearch.value
    const res = await diagnosticsAPI.listLogs(params)
    logsData.value = res.data || { logs: [], total: 0, page: 1, pages: 1 }
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Erreur chargement logs', 'error')
    logsData.value = { logs: [], total: 0, page: 1, pages: 1 }
  } finally {
    loading.value = false
  }
}

function viewLog(log: any) {
  selectedLog.value = log
  showDetailModal.value = true
}

async function resolveLog(id: string) {
  try {
    await diagnosticsAPI.resolveLog(id, {})
    store.addToast('Log résolu', 'success')
    await Promise.all([loadLogs(), loadStats()])
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Erreur', 'error')
  }
}

async function deleteLog(id: string) {
  if (!confirm('Supprimer ce log ?')) return
  try {
    await diagnosticsAPI.deleteLog(id)
    store.addToast('Log supprimé', 'success')
    await loadLogs()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Erreur suppression', 'error')
  }
}

async function bulkResolve() {
  if (!selectedIds.value.length) return
  submitting.value = true
  try {
    await diagnosticsAPI.bulkResolveLogs({ ids: selectedIds.value })
    store.addToast(`${selectedIds.value.length} logs résolus`, 'success')
    selectedIds.value = []
    await Promise.all([loadLogs(), loadStats()])
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Erreur', 'error')
  } finally {
    submitting.value = false
  }
}

const allSelected = computed(() =>
  logsData.value.logs.length > 0 && selectedIds.value.length === logsData.value.logs.length
)
function toggleSelectAll() {
  if (allSelected.value) {
    selectedIds.value = []
  } else {
    selectedIds.value = logsData.value.logs.map((l: any) => l.id)
  }
}

function openCreateModal() {
  createForm.value = defaultCreateForm()
  showCreateModal.value = true
}

async function submitCreateLog() {
  submitting.value = true
  try {
    await diagnosticsAPI.createLog(createForm.value)
    store.addToast('Log enregistré', 'success')
    showCreateModal.value = false
    await Promise.all([loadLogs(), loadStats()])
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Erreur', 'error')
  } finally {
    submitting.value = false
  }
}

function openPurgeModal() {
  purgeForm.value = { older_than_days: 90, severity: '', only_resolved: false }
  showPurgeModal.value = true
}

async function submitPurge() {
  submitting.value = true
  try {
    const res = await diagnosticsAPI.purgeLogs(purgeForm.value)
    store.addToast(`${res.data?.deleted || 0} logs supprimés`, 'success')
    showPurgeModal.value = false
    await Promise.all([loadLogs(), loadStats()])
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Erreur purge', 'error')
  } finally {
    submitting.value = false
  }
}

const maxHourly = computed(() => Math.max(...(stats.value.hourly_trend || []).map((h: any) => h.count), 1))

function severityBadge(s: string) {
  const m: Record<string, string> = {
    critical: 'bg-red-900/70 text-red-200 border border-red-700/50',
    error: 'bg-rose-900/60 text-rose-300 border border-rose-700/40',
    warning: 'bg-yellow-900/60 text-yellow-300 border border-yellow-700/40',
    info: 'bg-blue-900/60 text-blue-300 border border-blue-700/40',
    debug: 'bg-gray-800 text-gray-400 border border-gray-700'
  }
  return m[s] || m.debug
}
function severityBg(s: string) {
  const m: Record<string, string> = {
    critical: 'bg-red-700', error: 'bg-rose-700',
    warning: 'bg-yellow-700', info: 'bg-blue-700', debug: 'bg-gray-700'
  }
  return m[s] || 'bg-gray-700'
}
function severityIcon(s: string) {
  if (s === 'critical') return AlertOctagon
  if (s === 'error') return AlertCircle
  if (s === 'warning') return AlertTriangle
  if (s === 'info') return Info
  return Bug
}
function sourceBadge(s: string) {
  const m: Record<string, string> = {
    frontend_js: 'bg-yellow-900/50 text-yellow-300',
    backend_go: 'bg-cyan-900/50 text-cyan-300',
    database_sql: 'bg-purple-900/50 text-purple-300',
    api_http: 'bg-blue-900/50 text-blue-300',
    auth: 'bg-orange-900/50 text-orange-300',
    system: 'bg-gray-800 text-gray-400'
  }
  return m[s] || 'bg-gray-800 text-gray-400'
}
function sourceDot(s: string) {
  const m: Record<string, string> = {
    frontend_js: 'bg-yellow-400', backend_go: 'bg-cyan-400',
    database_sql: 'bg-purple-400', api_http: 'bg-blue-400',
    auth: 'bg-orange-400', system: 'bg-gray-500'
  }
  return m[s] || 'bg-gray-500'
}
function sourceBar(s: string) {
  const m: Record<string, string> = {
    frontend_js: 'bg-yellow-500', backend_go: 'bg-cyan-500',
    database_sql: 'bg-purple-500', api_http: 'bg-blue-500',
    auth: 'bg-orange-500', system: 'bg-gray-500'
  }
  return m[s] || 'bg-gray-500'
}
function sourceLabel(s: string) {
  const m: Record<string, string> = {
    frontend_js: 'Frontend JS', backend_go: 'Backend Go',
    database_sql: 'Database SQL', api_http: 'API HTTP',
    auth: 'Auth', background_job: 'Job', migration: 'Migration', system: 'System'
  }
  return m[s] || s
}
function httpStatusClass(s: number) {
  if (s >= 500) return 'bg-red-900/70 text-red-300'
  if (s >= 400) return 'bg-yellow-900/70 text-yellow-300'
  if (s >= 300) return 'bg-blue-900/70 text-blue-300'
  return 'bg-emerald-900/70 text-emerald-300'
}
function fmtRelative(d: string) {
  if (!d) return '—'
  const diff = Date.now() - new Date(d).getTime()
  const s = Math.floor(diff / 1000)
  if (s < 60) return `${s}s`
  const m = Math.floor(s / 60)
  if (m < 60) return `${m}m`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h`
  return `${Math.floor(h / 24)}j`
}
function fmtDateTime(d: string) {
  if (!d) return '—'
  return new Date(d).toLocaleString('fr-DZ', { day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit', second: '2-digit' })
}
function fmtHour(d: string) {
  if (!d) return '—'
  return new Date(d).toLocaleString('fr-DZ', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' })
}
function fmtHourShort(d: string) {
  if (!d) return ''
  return new Date(d).getHours() + 'h'
}

onMounted(() => {
  loadStats()
  loadLogs()
})
</script>

<style scoped>
.field {
  @apply w-full bg-gray-800 border border-gray-700 text-gray-200 rounded-xl px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-red-500 transition-all placeholder-gray-600;
}
.lbl {
  @apply block text-xs font-semibold text-gray-400 mb-1.5 uppercase tracking-wide;
}
</style>
