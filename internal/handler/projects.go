package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"nexus-erp/internal/middleware"
)

// ─── ProjectsHandler ──────────────────────────────────────────────────────────

type ProjectsHandler struct{ db *pgxpool.Pool }

// ═════════════════════════════════════════════════════════════════════════════
// PROJECTS
// ═════════════════════════════════════════════════════════════════════════════

func (h *ProjectsHandler) ListProjects(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	status := c.Query("status")
	search := c.Query("search")
	managerID := c.Query("manager_id")

	args := []interface{}{companyID}
	where := []string{"p.company_id = $1"}
	idx := 2

	if status != "" {
		where = append(where, "p.status = $"+itoa(idx))
		args = append(args, status)
		idx++
	}
	if search != "" {
		where = append(where, "(p.name ILIKE $"+itoa(idx)+" OR p.code ILIKE $"+itoa(idx)+")")
		args = append(args, "%"+search+"%")
		idx++
	}
	if managerID != "" {
		where = append(where, "p.manager_id = $"+itoa(idx))
		args = append(args, managerID)
		idx++
	}
	_ = idx

	sql := `
		SELECT
			p.id, p.code, p.name, p.status, p.start_date, p.end_date,
			p.budget, p.actual_cost, p.progress_pct, p.description, p.notes, p.color,
			p.customer_id, p.manager_id, p.created_at, p.updated_at,
			COALESCE(c.name,'')                               AS customer_name,
			COALESCE(e.first_name||' '||e.last_name,'')       AS manager_name,
			COALESCE(ts.total,0)                              AS total_tasks,
			COALESCE(ts.done,0)                               AS completed_tasks,
			COALESCE(th.hrs,0)                                AS total_hours,
			COALESCE(th.bhrs,0)                               AS billable_hours,
			COALESCE(ex.amt,0)                                AS total_expenses,
			COALESCE(ms.total,0)                              AS total_milestones,
			COALESCE(ms.done,0)                               AS completed_milestones
		FROM projects p
		LEFT JOIN customers c  ON c.id = p.customer_id
		LEFT JOIN employees e  ON e.id = p.manager_id
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS total, COUNT(*) FILTER (WHERE status='done') AS done
			FROM project_tasks WHERE project_id = p.id
		) ts ON TRUE
		LEFT JOIN LATERAL (
			SELECT COALESCE(SUM(hours),0) AS hrs,
			       COALESCE(SUM(hours) FILTER (WHERE billable),0) AS bhrs
			FROM timesheets WHERE project_id = p.id
		) th ON TRUE
		LEFT JOIN LATERAL (
			SELECT COALESCE(SUM(amount),0) AS amt
			FROM project_expenses WHERE project_id = p.id AND status IN ('approved','paid')
		) ex ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS total, COUNT(*) FILTER (WHERE status='completed') AS done
			FROM project_milestones WHERE project_id = p.id
		) ms ON TRUE
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY p.created_at DESC`

	rows, err := h.db.Query(ctx, sql, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var projects []map[string]interface{}
	for rows.Next() {
		var (
			id, code, name, status2, desc, notes, color, customerID, managerID2, customerName, managerName string
			startDate, endDate, createdAt, updatedAt                                                        interface{}
			budget, actualCost, totalHours, billableHours, totalExpenses                                    float64
			progressPct, totalTasks, completedTasks, totalMilestones, completedMilestones                   int
		)
		if err2 := rows.Scan(
			&id, &code, &name, &status2, &startDate, &endDate,
			&budget, &actualCost, &progressPct, &desc, &notes, &color,
			&customerID, &managerID2, &createdAt, &updatedAt,
			&customerName, &managerName,
			&totalTasks, &completedTasks, &totalHours, &billableHours,
			&totalExpenses, &totalMilestones, &completedMilestones,
		); err2 != nil {
			continue
		}
		projects = append(projects, map[string]interface{}{
			"id": id, "code": code, "name": name, "status": status2,
			"start_date": startDate, "end_date": endDate,
			"budget": budget, "actual_cost": actualCost, "progress_pct": progressPct,
			"description": desc, "notes": notes, "color": color,
			"customer_id": customerID, "manager_id": managerID2,
			"created_at": createdAt, "updated_at": updatedAt,
			"customer_name": customerName, "manager_name": managerName,
			"total_tasks": totalTasks, "completed_tasks": completedTasks,
			"total_hours": totalHours, "billable_hours": billableHours,
			"total_expenses": totalExpenses,
			"total_milestones": totalMilestones, "completed_milestones": completedMilestones,
		})
	}
	if projects == nil {
		projects = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, projects)
}

func (h *ProjectsHandler) GetProject(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	var (
		pID, code, name, status, desc, notes, color, customerID, managerID, customerName, managerName string
		startDate, endDate, createdAt, updatedAt                                                       interface{}
		budget, actualCost, totalHours, billableHours, totalExpenses                                   float64
		progressPct, totalTasks, completedTasks, totalMilestones, completedMilestones                  int
	)
	err := h.db.QueryRow(ctx, `
		SELECT
			p.id, p.code, p.name, p.status, p.start_date, p.end_date,
			p.budget, p.actual_cost, p.progress_pct, p.description, p.notes, p.color,
			p.customer_id, p.manager_id, p.created_at, p.updated_at,
			COALESCE(c.name,'')                               AS customer_name,
			COALESCE(e.first_name||' '||e.last_name,'')       AS manager_name,
			COALESCE(ts.total,0), COALESCE(ts.done,0),
			COALESCE(th.hrs,0), COALESCE(th.bhrs,0),
			COALESCE(ex.amt,0),
			COALESCE(ms.total,0), COALESCE(ms.done,0)
		FROM projects p
		LEFT JOIN customers c ON c.id = p.customer_id
		LEFT JOIN employees e ON e.id = p.manager_id
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS total, COUNT(*) FILTER (WHERE status='done') AS done
			FROM project_tasks WHERE project_id = p.id
		) ts ON TRUE
		LEFT JOIN LATERAL (
			SELECT COALESCE(SUM(hours),0) AS hrs, COALESCE(SUM(hours) FILTER (WHERE billable),0) AS bhrs
			FROM timesheets WHERE project_id = p.id
		) th ON TRUE
		LEFT JOIN LATERAL (
			SELECT COALESCE(SUM(amount),0) AS amt
			FROM project_expenses WHERE project_id = p.id AND status IN ('approved','paid')
		) ex ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS total, COUNT(*) FILTER (WHERE status='completed') AS done
			FROM project_milestones WHERE project_id = p.id
		) ms ON TRUE
		WHERE p.id = $1`, id).
		Scan(&pID, &code, &name, &status, &startDate, &endDate,
			&budget, &actualCost, &progressPct, &desc, &notes, &color,
			&customerID, &managerID, &createdAt, &updatedAt,
			&customerName, &managerName,
			&totalTasks, &completedTasks, &totalHours, &billableHours,
			&totalExpenses, &totalMilestones, &completedMilestones)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{
		"id": pID, "code": code, "name": name, "status": status,
		"start_date": startDate, "end_date": endDate,
		"budget": budget, "actual_cost": actualCost, "progress_pct": progressPct,
		"description": desc, "notes": notes, "color": color,
		"customer_id": customerID, "manager_id": managerID,
		"created_at": createdAt, "updated_at": updatedAt,
		"customer_name": customerName, "manager_name": managerName,
		"total_tasks": totalTasks, "completed_tasks": completedTasks,
		"total_hours": totalHours, "billable_hours": billableHours,
		"total_expenses": totalExpenses,
		"total_milestones": totalMilestones, "completed_milestones": completedMilestones,
	})
}

func (h *ProjectsHandler) CreateProject(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var body struct {
		Code        string   `json:"code"`
		Name        string   `json:"name"`
		CustomerID  *string  `json:"customer_id"`
		ManagerID   *string  `json:"manager_id"`
		StartDate   *string  `json:"start_date"`
		EndDate     *string  `json:"end_date"`
		Status      string   `json:"status"`
		Budget      float64  `json:"budget"`
		Description string   `json:"description"`
		Notes       string   `json:"notes"`
		Color       string   `json:"color"`
		AccountID   *string  `json:"account_id"`
		CostCenterID *string `json:"cost_center_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if body.Status == "" {
		body.Status = "planning"
	}
	if body.Color == "" {
		body.Color = "#6366f1"
	}

	// Auto-generate code if not provided
	if body.Code == "" {
		_ = h.db.QueryRow(ctx,
			`SELECT 'PRJ-'||LPAD((COALESCE(MAX(CAST(REGEXP_REPLACE(code,'[^0-9]','','g') AS INT)),0)+1)::TEXT,4,'0')
			 FROM projects WHERE company_id=$1 AND code ~ '^PRJ-[0-9]+$'`, companyID).
			Scan(&body.Code)
		if body.Code == "" {
			body.Code = "PRJ-0001"
		}
	}

	id := uuid.NewString()
	_, err := h.db.Exec(ctx, `
		INSERT INTO projects
			(id, company_id, code, name, customer_id, manager_id, start_date, end_date,
			 status, budget, description, notes, color, account_id, cost_center_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		id, companyID, body.Code, body.Name,
		nullableStr(body.CustomerID), nullableStr(body.ManagerID),
		nullableStr(body.StartDate), nullableStr(body.EndDate),
		body.Status, body.Budget, body.Description, body.Notes, body.Color,
		nullableStr(body.AccountID), nullableStr(body.CostCenterID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "code": body.Code, "name": body.Name})
}

func (h *ProjectsHandler) UpdateProject(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	var body struct {
		Name         string   `json:"name"`
		CustomerID   *string  `json:"customer_id"`
		ManagerID    *string  `json:"manager_id"`
		StartDate    *string  `json:"start_date"`
		EndDate      *string  `json:"end_date"`
		Status       string   `json:"status"`
		Budget       float64  `json:"budget"`
		ProgressPct  int      `json:"progress_pct"`
		Description  string   `json:"description"`
		Notes        string   `json:"notes"`
		Color        string   `json:"color"`
		AccountID    *string  `json:"account_id"`
		CostCenterID *string  `json:"cost_center_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.db.Exec(ctx, `
		UPDATE projects SET
			name=$1, customer_id=$2, manager_id=$3, start_date=$4, end_date=$5,
			status=$6, budget=$7, progress_pct=$8, description=$9, notes=$10, color=$11,
			account_id=$12, cost_center_id=$13, updated_at=NOW()
		WHERE id=$14`,
		body.Name, nullableStr(body.CustomerID), nullableStr(body.ManagerID),
		nullableStr(body.StartDate), nullableStr(body.EndDate),
		body.Status, body.Budget, body.ProgressPct, body.Description, body.Notes, body.Color,
		nullableStr(body.AccountID), nullableStr(body.CostCenterID), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *ProjectsHandler) DeleteProject(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `UPDATE projects SET status='cancelled', updated_at=NOW() WHERE id=$1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Project cancelled"})
}

func (h *ProjectsHandler) GetProjectDashboard(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var totalProjects, activePrj, completedPrj, onHoldPrj int
	var totalBudget, totalActual, totalHours float64
	_ = h.db.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status='active'),
			COUNT(*) FILTER (WHERE status='completed'),
			COUNT(*) FILTER (WHERE status='on_hold'),
			COALESCE(SUM(budget),0),
			COALESCE(SUM(actual_cost),0)
		FROM projects WHERE company_id=$1`, companyID).
		Scan(&totalProjects, &activePrj, &completedPrj, &onHoldPrj, &totalBudget, &totalActual)

	_ = h.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(ts.hours),0)
		FROM timesheets ts
		JOIN projects p ON p.id=ts.project_id
		WHERE p.company_id=$1`, companyID).Scan(&totalHours)

	var overdueTasks int
	_ = h.db.QueryRow(ctx, `
		SELECT COUNT(t.id) FROM project_tasks t
		JOIN projects p ON p.id=t.project_id
		WHERE p.company_id=$1 AND t.due_date < CURRENT_DATE AND t.status NOT IN ('done','cancelled')`,
		companyID).Scan(&overdueTasks)

	// Recent projects
	rows, _ := h.db.Query(ctx, `
		SELECT id, code, name, status, progress_pct, budget, actual_cost, end_date
		FROM projects WHERE company_id=$1
		ORDER BY updated_at DESC LIMIT 5`, companyID)
	var recent []map[string]interface{}
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var id2, code, name2, status2 string
			var prog int
			var bud, act float64
			var endDate interface{}
			if rows.Scan(&id2, &code, &name2, &status2, &prog, &bud, &act, &endDate) == nil {
				recent = append(recent, map[string]interface{}{
					"id": id2, "code": code, "name": name2, "status": status2,
					"progress_pct": prog, "budget": bud, "actual_cost": act, "end_date": endDate,
				})
			}
		}
	}
	if recent == nil {
		recent = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{
		"total_projects":   totalProjects,
		"active_projects":  activePrj,
		"completed":        completedPrj,
		"on_hold":          onHoldPrj,
		"total_budget":     totalBudget,
		"total_actual":     totalActual,
		"total_hours":      totalHours,
		"overdue_tasks":    overdueTasks,
		"budget_variance":  totalBudget - totalActual,
		"recent_projects":  recent,
	})
}

// ═════════════════════════════════════════════════════════════════════════════
// TASKS
// ═════════════════════════════════════════════════════════════════════════════

func (h *ProjectsHandler) ListTasks(c *gin.Context) {
	projectID := c.Param("id")
	ctx := context.Background()

	status := c.Query("status")
	assigneeID := c.Query("assignee_id")
	priority := c.Query("priority")

	args := []interface{}{projectID}
	where := []string{"t.project_id = $1"}
	idx := 2

	if status != "" {
		where = append(where, "t.status = $"+itoa(idx))
		args = append(args, status)
		idx++
	}
	if assigneeID != "" {
		where = append(where, "t.assignee_id = $"+itoa(idx))
		args = append(args, assigneeID)
		idx++
	}
	if priority != "" {
		where = append(where, "t.priority = $"+itoa(idx))
		args = append(args, priority)
		idx++
	}
	_ = idx

	rows, err := h.db.Query(ctx, `
		SELECT
			t.id, t.project_id, t.parent_id, t.title, t.description,
			t.assignee_id, t.status, t.priority, t.estimated_hours, t.actual_hours,
			t.start_date, t.due_date, t.completed_at, t.sort_order, t.color, t.tags,
			t.created_at, t.updated_at,
			COALESCE(e.first_name||' '||e.last_name,'') AS assignee_name,
			COALESCE(sub.sub_count,0), COALESCE(sub.sub_done,0),
			CASE WHEN t.due_date < CURRENT_DATE AND t.status NOT IN ('done','cancelled') THEN true ELSE false END AS is_overdue
		FROM project_tasks t
		LEFT JOIN employees e ON e.id = t.assignee_id
		LEFT JOIN LATERAL (
			SELECT COUNT(*) sub_count, COUNT(*) FILTER (WHERE status='done') sub_done
			FROM project_tasks WHERE parent_id = t.id
		) sub ON TRUE
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY t.sort_order ASC, t.due_date ASC NULLS LAST`, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var tasks []map[string]interface{}
	for rows.Next() {
		var (
			id, projID, title, desc, assigneeID2, status2, priority2, color, assigneeName string
			parentID, startDate, dueDate, completedAt, createdAt, updatedAt               interface{}
			estH, actH                                                                     float64
			sortOrder, subCount, subDone                                                   int
			isOverdue                                                                      bool
			tags                                                                           []string
		)
		if err2 := rows.Scan(
			&id, &projID, &parentID, &title, &desc,
			&assigneeID2, &status2, &priority2, &estH, &actH,
			&startDate, &dueDate, &completedAt, &sortOrder, &color, &tags,
			&createdAt, &updatedAt,
			&assigneeName, &subCount, &subDone, &isOverdue,
		); err2 != nil {
			continue
		}
		tasks = append(tasks, map[string]interface{}{
			"id": id, "project_id": projID, "parent_id": parentID,
			"title": title, "description": desc,
			"assignee_id": assigneeID2, "status": status2, "priority": priority2,
			"estimated_hours": estH, "actual_hours": actH,
			"start_date": startDate, "due_date": dueDate, "completed_at": completedAt,
			"sort_order": sortOrder, "color": color, "tags": tags,
			"created_at": createdAt, "updated_at": updatedAt,
			"assignee_name": assigneeName,
			"sub_task_count": subCount, "sub_task_done": subDone,
			"is_overdue": isOverdue,
		})
	}
	if tasks == nil {
		tasks = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, tasks)
}

func (h *ProjectsHandler) ListAllTasks(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	status := c.Query("status")
	assigneeID := c.Query("assignee_id")
	projectID := c.Query("project_id")

	args := []interface{}{companyID}
	where := []string{"p.company_id = $1"}
	idx := 2

	if status != "" && status != "all" {
		where = append(where, "t.status = $"+itoa(idx))
		args = append(args, status)
		idx++
	}
	if assigneeID != "" {
		where = append(where, "t.assignee_id = $"+itoa(idx))
		args = append(args, assigneeID)
		idx++
	}
	if projectID != "" {
		where = append(where, "t.project_id = $"+itoa(idx))
		args = append(args, projectID)
		idx++
	}
	_ = idx

	rows, err := h.db.Query(ctx, `
		SELECT
			t.id, t.project_id, t.title, t.status, t.priority,
			t.estimated_hours, t.actual_hours, t.due_date, t.start_date,
			t.assignee_id, t.sort_order, t.created_at,
			COALESCE(e.first_name||' '||e.last_name,'') AS assignee_name,
			p.name AS project_name, p.code AS project_code, p.color AS project_color,
			CASE WHEN t.due_date < CURRENT_DATE AND t.status NOT IN ('done','cancelled') THEN true ELSE false END AS is_overdue
		FROM project_tasks t
		JOIN projects p ON p.id = t.project_id
		LEFT JOIN employees e ON e.id = t.assignee_id
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY t.due_date ASC NULLS LAST, t.priority DESC`, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var tasks []map[string]interface{}
	for rows.Next() {
		var (
			id, projID, title, status2, priority2, assigneeID2, assigneeName, projName, projCode, projColor string
			estH, actH                                                                                       float64
			sortOrder                                                                                        int
			isOverdue                                                                                        bool
			dueDate, startDate, createdAt                                                                    interface{}
		)
		if rows.Scan(&id, &projID, &title, &status2, &priority2, &estH, &actH, &dueDate, &startDate,
			&assigneeID2, &sortOrder, &createdAt, &assigneeName, &projName, &projCode, &projColor, &isOverdue) != nil {
			continue
		}
		tasks = append(tasks, map[string]interface{}{
			"id": id, "project_id": projID, "title": title,
			"status": status2, "priority": priority2,
			"estimated_hours": estH, "actual_hours": actH,
			"due_date": dueDate, "start_date": startDate,
			"assignee_id": assigneeID2, "assignee_name": assigneeName,
			"project_name": projName, "project_code": projCode, "project_color": projColor,
			"sort_order": sortOrder, "created_at": createdAt, "is_overdue": isOverdue,
		})
	}
	if tasks == nil {
		tasks = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, tasks)
}

func (h *ProjectsHandler) GetTask(c *gin.Context) {
	taskID := c.Param("taskId")
	ctx := context.Background()
	var (
		id, projID, title, desc, assigneeID2, status2, priority2, assigneeName, projName string
		parentID, startDate, dueDate, completedAt, createdAt, updatedAt                   interface{}
		estH, actH                                                                        float64
		sortOrder                                                                         int
	)
	err := h.db.QueryRow(ctx, `
		SELECT t.id, t.project_id, t.parent_id, t.title, t.description,
		       t.assignee_id, t.status, t.priority, t.estimated_hours, t.actual_hours,
		       t.start_date, t.due_date, t.completed_at, t.sort_order, t.created_at, t.updated_at,
		       COALESCE(e.first_name||' '||e.last_name,'') AS assignee_name,
		       p.name AS project_name
		FROM project_tasks t
		LEFT JOIN employees e ON e.id=t.assignee_id
		LEFT JOIN projects p ON p.id=t.project_id
		WHERE t.id=$1`, taskID).
		Scan(&id, &projID, &parentID, &title, &desc, &assigneeID2, &status2, &priority2,
			&estH, &actH, &startDate, &dueDate, &completedAt, &sortOrder, &createdAt, &updatedAt,
			&assigneeName, &projName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{
		"id": id, "project_id": projID, "parent_id": parentID,
		"title": title, "description": desc,
		"assignee_id": assigneeID2, "status": status2, "priority": priority2,
		"estimated_hours": estH, "actual_hours": actH,
		"start_date": startDate, "due_date": dueDate, "completed_at": completedAt,
		"sort_order": sortOrder, "created_at": createdAt, "updated_at": updatedAt,
		"assignee_name": assigneeName, "project_name": projName,
	})
}

func (h *ProjectsHandler) CreateTask(c *gin.Context) {
	projectID := c.Param("id")
	ctx := context.Background()
	var body struct {
		ParentID       *string `json:"parent_id"`
		Title          string  `json:"title"`
		Description    string  `json:"description"`
		AssigneeID     *string `json:"assignee_id"`
		Status         string  `json:"status"`
		Priority       string  `json:"priority"`
		EstimatedHours float64 `json:"estimated_hours"`
		StartDate      *string `json:"start_date"`
		DueDate        *string `json:"due_date"`
		SortOrder      int     `json:"sort_order"`
		Color          string  `json:"color"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}
	if body.Status == "" {
		body.Status = "todo"
	}
	if body.Priority == "" {
		body.Priority = "medium"
	}
	id := uuid.NewString()
	_, err := h.db.Exec(ctx, `
		INSERT INTO project_tasks
			(id, project_id, parent_id, title, description, assignee_id,
			 status, priority, estimated_hours, start_date, due_date, sort_order, color)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		id, projectID, nullableStr(body.ParentID), body.Title, body.Description,
		nullableStr(body.AssigneeID), body.Status, body.Priority, body.EstimatedHours,
		nullableStr(body.StartDate), nullableStr(body.DueDate), body.SortOrder, body.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "project_id": projectID, "title": body.Title})
}

func (h *ProjectsHandler) UpdateTask(c *gin.Context) {
	taskID := c.Param("taskId")
	ctx := context.Background()
	var body struct {
		Title          string  `json:"title"`
		Description    string  `json:"description"`
		AssigneeID     *string `json:"assignee_id"`
		Status         string  `json:"status"`
		Priority       string  `json:"priority"`
		EstimatedHours float64 `json:"estimated_hours"`
		ActualHours    float64 `json:"actual_hours"`
		StartDate      *string `json:"start_date"`
		DueDate        *string `json:"due_date"`
		SortOrder      int     `json:"sort_order"`
		Color          string  `json:"color"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var completedAt interface{}
	if body.Status == "done" {
		now := time.Now()
		completedAt = now
	}
	_, err := h.db.Exec(ctx, `
		UPDATE project_tasks SET
			title=$1, description=$2, assignee_id=$3, status=$4, priority=$5,
			estimated_hours=$6, actual_hours=$7, start_date=$8, due_date=$9,
			sort_order=$10, color=$11, completed_at=$12, updated_at=NOW()
		WHERE id=$13`,
		body.Title, body.Description, nullableStr(body.AssigneeID), body.Status, body.Priority,
		body.EstimatedHours, body.ActualHours,
		nullableStr(body.StartDate), nullableStr(body.DueDate),
		body.SortOrder, body.Color, completedAt, taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Recalculate project progress
	var projID string
	if h.db.QueryRow(ctx, `SELECT project_id FROM project_tasks WHERE id=$1`, taskID).Scan(&projID) == nil {
		_, _ = h.db.Exec(ctx, `SELECT update_project_progress($1)`, projID)
	}
	c.JSON(http.StatusOK, gin.H{"id": taskID})
}

func (h *ProjectsHandler) DeleteTask(c *gin.Context) {
	taskID := c.Param("taskId")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `UPDATE project_tasks SET status='cancelled', updated_at=NOW() WHERE id=$1`, taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task cancelled"})
}

// ═════════════════════════════════════════════════════════════════════════════
// MILESTONES
// ═════════════════════════════════════════════════════════════════════════════

func (h *ProjectsHandler) ListMilestones(c *gin.Context) {
	projectID := c.Param("id")
	ctx := context.Background()
	rows, err := h.db.Query(ctx, `
		SELECT
			m.id, m.project_id, m.title, m.description, m.due_date,
			m.completed_at, m.status, m.owner_id, m.progress_pct, m.sort_order,
			m.created_at, m.updated_at,
			COALESCE(e.first_name||' '||e.last_name,'') AS owner_name
		FROM project_milestones m
		LEFT JOIN employees e ON e.id = m.owner_id
		WHERE m.project_id=$1
		ORDER BY m.sort_order ASC, m.due_date ASC`, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var (
			id, projID, title, desc, status2, ownerID, ownerName string
			dueDate, completedAt, createdAt, updatedAt            interface{}
			progressPct, sortOrder                                 int
		)
		if rows.Scan(&id, &projID, &title, &desc, &dueDate, &completedAt, &status2,
			&ownerID, &progressPct, &sortOrder, &createdAt, &updatedAt, &ownerName) != nil {
			continue
		}
		list = append(list, map[string]interface{}{
			"id": id, "project_id": projID, "title": title, "description": desc,
			"due_date": dueDate, "completed_at": completedAt, "status": status2,
			"owner_id": ownerID, "owner_name": ownerName,
			"progress_pct": progressPct, "sort_order": sortOrder,
			"created_at": createdAt, "updated_at": updatedAt,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *ProjectsHandler) CreateMilestone(c *gin.Context) {
	projectID := c.Param("id")
	ctx := context.Background()
	var body struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		DueDate     string  `json:"due_date"`
		OwnerID     *string `json:"owner_id"`
		Status      string  `json:"status"`
		ProgressPct int     `json:"progress_pct"`
		SortOrder   int     `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Title == "" || body.DueDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title and due_date are required"})
		return
	}
	if body.Status == "" {
		body.Status = "pending"
	}
	id := uuid.NewString()
	_, err := h.db.Exec(ctx, `
		INSERT INTO project_milestones
			(id, project_id, title, description, due_date, owner_id, status, progress_pct, sort_order)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		id, projectID, body.Title, body.Description, body.DueDate,
		nullableStr(body.OwnerID), body.Status, body.ProgressPct, body.SortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *ProjectsHandler) UpdateMilestone(c *gin.Context) {
	milestoneID := c.Param("milestoneId")
	ctx := context.Background()
	var body struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		DueDate     string  `json:"due_date"`
		OwnerID     *string `json:"owner_id"`
		Status      string  `json:"status"`
		ProgressPct int     `json:"progress_pct"`
		SortOrder   int     `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var completedAt interface{}
	if body.Status == "completed" {
		now := time.Now()
		completedAt = now
	}
	_, err := h.db.Exec(ctx, `
		UPDATE project_milestones SET
			title=$1, description=$2, due_date=$3, owner_id=$4, status=$5,
			progress_pct=$6, sort_order=$7, completed_at=$8, updated_at=NOW()
		WHERE id=$9`,
		body.Title, body.Description, body.DueDate, nullableStr(body.OwnerID), body.Status,
		body.ProgressPct, body.SortOrder, completedAt, milestoneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": milestoneID})
}

func (h *ProjectsHandler) DeleteMilestone(c *gin.Context) {
	id := c.Param("milestoneId")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `DELETE FROM project_milestones WHERE id=$1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

// ═════════════════════════════════════════════════════════════════════════════
// TIMESHEETS
// ═════════════════════════════════════════════════════════════════════════════

func (h *ProjectsHandler) ListTimesheets(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	projectID := c.Query("project_id")
	employeeID := c.Query("employee_id")
	month := c.Query("month")
	year := c.Query("year")

	args := []interface{}{companyID}
	where := []string{"ts.company_id = $1"}
	idx := 2

	if projectID != "" {
		where = append(where, "ts.project_id = $"+itoa(idx))
		args = append(args, projectID)
		idx++
	}
	if employeeID != "" {
		where = append(where, "ts.employee_id = $"+itoa(idx))
		args = append(args, employeeID)
		idx++
	}
	if month != "" {
		where = append(where, "EXTRACT(MONTH FROM ts.date) = $"+itoa(idx))
		args = append(args, month)
		idx++
	}
	if year != "" {
		where = append(where, "EXTRACT(YEAR FROM ts.date) = $"+itoa(idx))
		args = append(args, year)
		idx++
	}
	_ = idx

	rows, err := h.db.Query(ctx, `
		SELECT
			ts.id, ts.employee_id, ts.project_id, ts.task_id, ts.date,
			ts.hours, ts.hourly_rate, ts.hours * ts.hourly_rate AS line_amount,
			ts.description, ts.billable, ts.billed, ts.approved, ts.created_at,
			COALESCE(e.first_name||' '||e.last_name,'') AS employee_name,
			COALESCE(e.employee_number,'')               AS employee_number,
			p.name AS project_name, p.code AS project_code,
			COALESCE(t.title,'')                         AS task_title
		FROM timesheets ts
		LEFT JOIN employees     e ON e.id = ts.employee_id
		LEFT JOIN projects      p ON p.id = ts.project_id
		LEFT JOIN project_tasks t ON t.id = ts.task_id
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY ts.date DESC, ts.created_at DESC
		LIMIT 500`, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var (
			id, empID, projID, empName, empNum, projName, projCode, taskTitle, desc string
			taskID, date, createdAt                                                  interface{}
			hours, rate, lineAmt                                                     float64
			billable, billed, approved                                               bool
		)
		if rows.Scan(&id, &empID, &projID, &taskID, &date, &hours, &rate, &lineAmt, &desc,
			&billable, &billed, &approved, &createdAt,
			&empName, &empNum, &projName, &projCode, &taskTitle) != nil {
			continue
		}
		list = append(list, map[string]interface{}{
			"id": id, "employee_id": empID, "project_id": projID, "task_id": taskID,
			"date": date, "hours": hours, "hourly_rate": rate, "line_amount": lineAmt,
			"description": desc, "billable": billable, "billed": billed, "approved": approved,
			"created_at": createdAt, "employee_name": empName, "employee_number": empNum,
			"project_name": projName, "project_code": projCode, "task_title": taskTitle,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *ProjectsHandler) CreateTimesheet(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	var body struct {
		EmployeeID  string  `json:"employee_id"`
		ProjectID   string  `json:"project_id"`
		TaskID      *string `json:"task_id"`
		Date        string  `json:"date"`
		Hours       float64 `json:"hours"`
		HourlyRate  float64 `json:"hourly_rate"`
		Description string  `json:"description"`
		Billable    bool    `json:"billable"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.EmployeeID == "" || body.ProjectID == "" || body.Date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id, project_id and date are required"})
		return
	}
	id := uuid.NewString()
	_, err := h.db.Exec(ctx, `
		INSERT INTO timesheets
			(id, company_id, employee_id, project_id, task_id, date, hours, hourly_rate, description, billable)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, companyID, body.EmployeeID, body.ProjectID,
		nullableStr(body.TaskID), body.Date, body.Hours, body.HourlyRate,
		body.Description, body.Billable)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Update project actual_cost
	_, _ = h.db.Exec(ctx,
		`UPDATE projects SET actual_cost = actual_cost + $1, updated_at=NOW() WHERE id=$2`,
		body.Hours*body.HourlyRate, body.ProjectID)
	// Update task actual_hours
	if body.TaskID != nil && *body.TaskID != "" {
		_, _ = h.db.Exec(ctx,
			`UPDATE project_tasks SET actual_hours = actual_hours + $1, updated_at=NOW() WHERE id=$2`,
			body.Hours, *body.TaskID)
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *ProjectsHandler) UpdateTimesheet(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	var body struct {
		Hours       float64 `json:"hours"`
		HourlyRate  float64 `json:"hourly_rate"`
		Description string  `json:"description"`
		Billable    bool    `json:"billable"`
		Approved    bool    `json:"approved"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.db.Exec(ctx, `
		UPDATE timesheets SET hours=$1, hourly_rate=$2, description=$3, billable=$4, approved=$5
		WHERE id=$6`,
		body.Hours, body.HourlyRate, body.Description, body.Billable, body.Approved, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *ProjectsHandler) DeleteTimesheet(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `DELETE FROM timesheets WHERE id=$1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

// ═════════════════════════════════════════════════════════════════════════════
// EXPENSES
// ═════════════════════════════════════════════════════════════════════════════

func (h *ProjectsHandler) ListExpenses(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	projectID := c.Query("project_id")
	status := c.Query("status")
	employeeID := c.Query("employee_id")

	args := []interface{}{companyID}
	where := []string{"ex.company_id = $1"}
	idx := 2

	if projectID != "" {
		where = append(where, "ex.project_id = $"+itoa(idx))
		args = append(args, projectID)
		idx++
	}
	if status != "" {
		where = append(where, "ex.status = $"+itoa(idx))
		args = append(args, status)
		idx++
	}
	if employeeID != "" {
		where = append(where, "ex.employee_id = $"+itoa(idx))
		args = append(args, employeeID)
		idx++
	}
	_ = idx

	rows, err := h.db.Query(ctx, `
		SELECT
			ex.id, ex.project_id, ex.task_id, ex.employee_id, ex.category,
			ex.description, ex.amount, ex.currency, ex.expense_date,
			ex.receipt_url, ex.status, ex.is_billable, ex.billed, ex.created_at,
			COALESCE(e.first_name||' '||e.last_name,'') AS employee_name,
			p.name AS project_name, p.code AS project_code,
			COALESCE(t.title,'') AS task_title
		FROM project_expenses ex
		LEFT JOIN employees     e ON e.id = ex.employee_id
		LEFT JOIN projects      p ON p.id = ex.project_id
		LEFT JOIN project_tasks t ON t.id = ex.task_id
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY ex.expense_date DESC, ex.created_at DESC`, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var (
			id, projID, empID, category, desc, currency, receiptURL, status2, empName, projName, projCode, taskTitle string
			taskID, expDate, createdAt                                                                                 interface{}
			amount                                                                                                     float64
			isBillable, billed                                                                                         bool
		)
		if rows.Scan(&id, &projID, &taskID, &empID, &category, &desc, &amount, &currency,
			&expDate, &receiptURL, &status2, &isBillable, &billed, &createdAt,
			&empName, &projName, &projCode, &taskTitle) != nil {
			continue
		}
		list = append(list, map[string]interface{}{
			"id": id, "project_id": projID, "task_id": taskID, "employee_id": empID,
			"category": category, "description": desc, "amount": amount, "currency": currency,
			"expense_date": expDate, "receipt_url": receiptURL, "status": status2,
			"is_billable": isBillable, "billed": billed, "created_at": createdAt,
			"employee_name": empName, "project_name": projName, "project_code": projCode,
			"task_title": taskTitle,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *ProjectsHandler) CreateExpense(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	var body struct {
		ProjectID   string  `json:"project_id"`
		TaskID      *string `json:"task_id"`
		EmployeeID  string  `json:"employee_id"`
		Category    string  `json:"category"`
		Description string  `json:"description"`
		Amount      float64 `json:"amount"`
		Currency    string  `json:"currency"`
		ExpenseDate string  `json:"expense_date"`
		ReceiptURL  string  `json:"receipt_url"`
		Status      string  `json:"status"`
		IsBillable  bool    `json:"is_billable"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.ProjectID == "" || body.EmployeeID == "" || body.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id, employee_id and description are required"})
		return
	}
	if body.Status == "" {
		body.Status = "draft"
	}
	if body.Currency == "" {
		body.Currency = "DZD"
	}
	if body.Category == "" {
		body.Category = "other"
	}
	if body.ExpenseDate == "" {
		body.ExpenseDate = time.Now().Format("2006-01-02")
	}
	id := uuid.NewString()
	_, err := h.db.Exec(ctx, `
		INSERT INTO project_expenses
			(id, project_id, task_id, employee_id, company_id, category, description,
			 amount, currency, expense_date, receipt_url, status, is_billable)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		id, body.ProjectID, nullableStr(body.TaskID), body.EmployeeID, companyID,
		body.Category, body.Description, body.Amount, body.Currency, body.ExpenseDate,
		body.ReceiptURL, body.Status, body.IsBillable)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *ProjectsHandler) UpdateExpense(c *gin.Context) {
	id := c.Param("expenseId")
	ctx := context.Background()
	var body struct {
		Category    string  `json:"category"`
		Description string  `json:"description"`
		Amount      float64 `json:"amount"`
		Currency    string  `json:"currency"`
		ExpenseDate string  `json:"expense_date"`
		ReceiptURL  string  `json:"receipt_url"`
		Status      string  `json:"status"`
		IsBillable  bool    `json:"is_billable"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var approvedAt interface{}
	if body.Status == "approved" {
		now := time.Now()
		approvedAt = now
	}
	_, err := h.db.Exec(ctx, `
		UPDATE project_expenses SET
			category=$1, description=$2, amount=$3, currency=$4,
			expense_date=$5, receipt_url=$6, status=$7, is_billable=$8,
			approved_at=$9, updated_at=NOW()
		WHERE id=$10`,
		body.Category, body.Description, body.Amount, body.Currency,
		body.ExpenseDate, body.ReceiptURL, body.Status, body.IsBillable,
		approvedAt, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *ProjectsHandler) DeleteExpense(c *gin.Context) {
	id := c.Param("expenseId")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `DELETE FROM project_expenses WHERE id=$1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

// ═════════════════════════════════════════════════════════════════════════════
// PLANNING SLOTS
// ═════════════════════════════════════════════════════════════════════════════

func (h *ProjectsHandler) ListPlanningSlots(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	projectID := c.Query("project_id")
	employeeID := c.Query("employee_id")
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")

	args := []interface{}{companyID}
	where := []string{"ps.company_id = $1"}
	idx := 2

	if projectID != "" {
		where = append(where, "ps.project_id = $"+itoa(idx))
		args = append(args, projectID)
		idx++
	}
	if employeeID != "" {
		where = append(where, "ps.employee_id = $"+itoa(idx))
		args = append(args, employeeID)
		idx++
	}
	if fromDate != "" {
		where = append(where, "ps.planned_date >= $"+itoa(idx))
		args = append(args, fromDate)
		idx++
	}
	if toDate != "" {
		where = append(where, "ps.planned_date <= $"+itoa(idx))
		args = append(args, toDate)
		idx++
	}
	_ = idx

	rows, err := h.db.Query(ctx, `
		SELECT
			ps.id, ps.project_id, ps.task_id, ps.employee_id,
			ps.planned_date, ps.planned_hours, ps.notes, ps.created_at,
			COALESCE(e.first_name||' '||e.last_name,'') AS employee_name,
			p.name AS project_name, p.code AS project_code, p.color AS project_color,
			COALESCE(t.title,'') AS task_title
		FROM planning_slots ps
		LEFT JOIN employees     e ON e.id = ps.employee_id
		LEFT JOIN projects      p ON p.id = ps.project_id
		LEFT JOIN project_tasks t ON t.id = ps.task_id
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY ps.planned_date ASC`, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var (
			id, projID, empID, empName, projName, projCode, projColor, taskTitle, notes string
			taskID, plannedDate, createdAt                                               interface{}
			plannedHours                                                                 float64
		)
		if rows.Scan(&id, &projID, &taskID, &empID, &plannedDate, &plannedHours, &notes, &createdAt,
			&empName, &projName, &projCode, &projColor, &taskTitle) != nil {
			continue
		}
		list = append(list, map[string]interface{}{
			"id": id, "project_id": projID, "task_id": taskID, "employee_id": empID,
			"planned_date": plannedDate, "planned_hours": plannedHours, "notes": notes,
			"created_at": createdAt, "employee_name": empName,
			"project_name": projName, "project_code": projCode, "project_color": projColor,
			"task_title": taskTitle,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *ProjectsHandler) UpsertPlanningSlot(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	var body struct {
		ProjectID    string  `json:"project_id"`
		TaskID       *string `json:"task_id"`
		EmployeeID   string  `json:"employee_id"`
		PlannedDate  string  `json:"planned_date"`
		PlannedHours float64 `json:"planned_hours"`
		Notes        string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.ProjectID == "" || body.EmployeeID == "" || body.PlannedDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id, employee_id and planned_date are required"})
		return
	}
	id := uuid.NewString()
	_, err := h.db.Exec(ctx, `
		INSERT INTO planning_slots
			(id, project_id, task_id, employee_id, company_id, planned_date, planned_hours, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (task_id, employee_id, planned_date)
		DO UPDATE SET planned_hours=$7, notes=$8`,
		id, body.ProjectID, nullableStr(body.TaskID), body.EmployeeID, companyID,
		body.PlannedDate, body.PlannedHours, body.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *ProjectsHandler) DeletePlanningSlot(c *gin.Context) {
	id := c.Param("slotId")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `DELETE FROM planning_slots WHERE id=$1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

// ═════════════════════════════════════════════════════════════════════════════
// REPORTS
// ═════════════════════════════════════════════════════════════════════════════

func (h *ProjectsHandler) GetProjectCosts(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	var budget, actualCost float64
	var name, code, status string
	var progressPct int
	_ = h.db.QueryRow(ctx,
		`SELECT name, code, status, budget, actual_cost, progress_pct FROM projects WHERE id=$1`, id).
		Scan(&name, &code, &status, &budget, &actualCost, &progressPct)

	var laborCost, expenseCost, billableHours float64
	_ = h.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(hours*hourly_rate),0), COALESCE(SUM(hours) FILTER (WHERE billable),0)
		 FROM timesheets WHERE project_id=$1`, id).Scan(&laborCost, &billableHours)

	_ = h.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount),0) FROM project_expenses WHERE project_id=$1 AND status IN ('approved','paid')`, id).
		Scan(&expenseCost)

	// Breakdown by employee
	rows, _ := h.db.Query(ctx, `
		SELECT COALESCE(e.first_name||' '||e.last_name,'Unknown') AS emp_name,
		       COALESCE(SUM(ts.hours),0) AS hours,
		       COALESCE(SUM(ts.hours*ts.hourly_rate),0) AS amount
		FROM timesheets ts
		LEFT JOIN employees e ON e.id=ts.employee_id
		WHERE ts.project_id=$1
		GROUP BY e.id, e.first_name, e.last_name
		ORDER BY amount DESC`, id)
	var byEmployee []map[string]interface{}
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var empName string
			var hrs, amt float64
			if rows.Scan(&empName, &hrs, &amt) == nil {
				byEmployee = append(byEmployee, map[string]interface{}{
					"employee_name": empName, "hours": hrs, "amount": amt,
				})
			}
		}
	}
	if byEmployee == nil {
		byEmployee = []map[string]interface{}{}
	}

	// Breakdown by task
	rows2, _ := h.db.Query(ctx, `
		SELECT t.title, COALESCE(SUM(ts.hours),0) AS hours
		FROM timesheets ts
		JOIN project_tasks t ON t.id=ts.task_id
		WHERE ts.project_id=$1
		GROUP BY t.id, t.title
		ORDER BY hours DESC LIMIT 10`, id)
	var byTask []map[string]interface{}
	if rows2 != nil {
		defer rows2.Close()
		for rows2.Next() {
			var title string
			var hrs float64
			if rows2.Scan(&title, &hrs) == nil {
				byTask = append(byTask, map[string]interface{}{"title": title, "hours": hrs})
			}
		}
	}
	if byTask == nil {
		byTask = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{
		"project_name": name, "project_code": code, "status": status,
		"budget": budget, "actual_cost": actualCost, "progress_pct": progressPct,
		"labor_cost": laborCost, "expense_cost": expenseCost,
		"total_cost": laborCost + expenseCost,
		"billable_hours": billableHours,
		"budget_variance": budget - (laborCost + expenseCost),
		"budget_used_pct": func() float64 {
			if budget == 0 {
				return 0
			}
			return ((laborCost + expenseCost) / budget) * 100
		}(),
		"by_employee": byEmployee,
		"by_task":     byTask,
	})
}

func (h *ProjectsHandler) GetProjectsReport(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT
			p.id, p.code, p.name, p.status, p.start_date, p.end_date,
			p.budget, p.actual_cost, p.progress_pct,
			COALESCE(ts.hrs,0) AS total_hours,
			COALESCE(ts.bhrs,0) AS billable_hours,
			COALESCE(ex.amt,0) AS total_expenses,
			COALESCE(tk.total,0) AS total_tasks,
			COALESCE(tk.done,0) AS done_tasks,
			COALESCE(ms.total,0) AS total_milestones,
			COALESCE(ms.done,0) AS done_milestones
		FROM projects p
		LEFT JOIN LATERAL (
			SELECT COALESCE(SUM(hours),0) AS hrs,
			       COALESCE(SUM(hours) FILTER (WHERE billable),0) AS bhrs
			FROM timesheets WHERE project_id=p.id
		) ts ON TRUE
		LEFT JOIN LATERAL (
			SELECT COALESCE(SUM(amount),0) AS amt
			FROM project_expenses WHERE project_id=p.id AND status IN ('approved','paid')
		) ex ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS total, COUNT(*) FILTER (WHERE status='done') AS done
			FROM project_tasks WHERE project_id=p.id
		) tk ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS total, COUNT(*) FILTER (WHERE status='completed') AS done
			FROM project_milestones WHERE project_id=p.id
		) ms ON TRUE
		WHERE p.company_id=$1
		ORDER BY p.created_at DESC`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	var totalBudget, totalActual, totalHoursAll, totalExpAll float64
	for rows.Next() {
		var (
			id, code, name, status string
			startDate, endDate     interface{}
			bud, act, hrs, bhrs, exAmt float64
			prog, tkTotal, tkDone, msTotal, msDone int
		)
		if rows.Scan(&id, &code, &name, &status, &startDate, &endDate, &bud, &act, &prog,
			&hrs, &bhrs, &exAmt, &tkTotal, &tkDone, &msTotal, &msDone) != nil {
			continue
		}
		totalBudget += bud
		totalActual += act
		totalHoursAll += hrs
		totalExpAll += exAmt
		list = append(list, map[string]interface{}{
			"id": id, "code": code, "name": name, "status": status,
			"start_date": startDate, "end_date": endDate,
			"budget": bud, "actual_cost": act, "progress_pct": prog,
			"total_hours": hrs, "billable_hours": bhrs, "total_expenses": exAmt,
			"total_tasks": tkTotal, "done_tasks": tkDone,
			"total_milestones": msTotal, "done_milestones": msDone,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, gin.H{
		"projects":      list,
		"total_budget":  totalBudget,
		"total_actual":  totalActual,
		"total_hours":   totalHoursAll,
		"total_expenses": totalExpAll,
		"variance":      totalBudget - totalActual,
	})
}

// ═════════════════════════════════════════════════════════════════════════════
// LEGACY (kept for backward compat)
// ═════════════════════════════════════════════════════════════════════════════

func (h *ProjectsHandler) GetProjectCostsLegacy(c *gin.Context) {
	h.GetProjectCosts(c)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}

func nullableStr(s *string) interface{} {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}
