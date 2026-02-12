class SalesTrackerApp {
  constructor() {
    this.currentPage = 1;
    this.itemsPerPage = 25;
    this.totalItems = 0;
    this.currentTab = "records";
    this.currentGroup = "none";
    this.itemsCache = [];
    this.categoriesCache = new Set();
    this.initEventListeners();
    this.loadItems();
    this.loadCategories();
    this.setupCharts();
  }
  initEventListeners() {
    document.querySelectorAll(".tab-btn").forEach((btn) => {
      btn.addEventListener("click", () => {
        const tabName = btn.dataset.tab;
        this.switchTab(tabName);
      });
    });
    document
      .getElementById("addItemForm")
      .addEventListener("submit", (e) => this.handleAddItem(e));
    document
      .getElementById("editItemForm")
      .addEventListener("submit", (e) => this.handleUpdateItem(e));
    document
      .getElementById("analyticsForm")
      .addEventListener("submit", (e) => this.handleGetAnalytics(e));
    document
      .getElementById("generateReport")
      .addEventListener("click", () => this.generateReport());
    document
      .getElementById("closeEditModal")
      .addEventListener("click", () => this.closeEditModal());
    document.getElementById("editModal").addEventListener("click", (e) => {
      if (e.target === document.getElementById("editModal"))
        this.closeEditModal();
    });
    document
      .getElementById("prevPage")
      .addEventListener("click", () => this.changePage(-1));
    document
      .getElementById("nextPage")
      .addEventListener("click", () => this.changePage(1));
    document.getElementById("itemsPerPage").addEventListener("change", (e) => {
      this.itemsPerPage = parseInt(e.target.value);
      this.currentPage = 1;
      this.renderItemsTable();
    });
    document
      .getElementById("sortOrder")
      .addEventListener("change", () => this.renderItemsTable());
    document
      .getElementById("applyFilters")
      .addEventListener("click", () => this.applyFilters());
    document
      .getElementById("globalSearch")
      .addEventListener("input", (e) => this.globalSearch(e.target.value));
    document
      .getElementById("filterType")
      .addEventListener("change", () => this.applyFilters());
    document
      .getElementById("filterCategory")
      .addEventListener("change", () => this.applyFilters());
    document
      .getElementById("filterPeriod")
      .addEventListener("change", () => this.applyFilters());
    document
      .getElementById("exportBtn")
      .addEventListener("click", () => this.exportAllToCSV());
    document
      .getElementById("exportAnalyticsCsv")
      .addEventListener("click", () => this.exportAnalyticsToCSV());
    document
      .getElementById("deleteItemBtn")
      .addEventListener("click", () => this.confirmDeleteItem());
    document.getElementById("toast").addEventListener("click", (e) => {
      if (e.target.id === "toast") this.hideToast();
    });
    document.addEventListener("keydown", (e) => {
      if (
        e.key === "Escape" &&
        !document.getElementById("editModal").classList.contains("hidden")
      ) {
        this.closeEditModal();
      }
      if (e.ctrlKey && e.key === "n") {
        e.preventDefault();
        document.getElementById("itemAmount").focus();
      }
    });
  }
  async loadItems() {
    try {
      const response = await fetch(
        `/items?page=${this.currentPage}&limit=${this.itemsPerPage}`,
      );
      if (!response.ok) throw new Error("Ошибка загрузки записей");
      const data = await response.json();
      this.itemsCache = data.items || [];
      this.totalItems = data.total || this.itemsCache.length;
      this.updateCategoriesDropdown();
      this.renderItemsTable();
      this.showSuccess("Записи загружены", `${this.itemsCache.length} записей`);
    } catch (error) {
      this.showError("Ошибка загрузки записей", error.message);
      console.error("Load items error:", error);
    }
  }
  async loadCategories() {
    try {
      this.itemsCache.forEach((item) => {
        if (item.category) this.categoriesCache.add(item.category);
      });
      const filterSelect = document.getElementById("filterCategory");
      filterSelect.innerHTML = '<option value="">Все категории</option>';
      Array.from(this.categoriesCache)
        .sort()
        .forEach((category) => {
          const option = document.createElement("option");
          option.value = category;
          option.textContent = category;
          filterSelect.appendChild(option);
        });
    } catch (error) {
      console.error("Load categories error:", error);
    }
  }
  updateCategoriesDropdown() {
    this.categoriesCache.clear();
    this.itemsCache.forEach((item) => {
      if (item.category) this.categoriesCache.add(item.category);
    });
    this.loadCategories();
  }
  renderItemsTable() {
    const tbody = document.getElementById("itemsTableBody");
    const startIdx = (this.currentPage - 1) * this.itemsPerPage;
    const endIdx = Math.min(
      startIdx + this.itemsPerPage,
      this.itemsCache.length,
    );
    const pageItems = this.itemsCache.slice(startIdx, endIdx);
    const sortOrder = document.getElementById("sortOrder").value;
    pageItems.sort((a, b) => {
      switch (sortOrder) {
        case "date_desc":
          return new Date(b.date) - new Date(a.date);
        case "date_asc":
          return new Date(a.date) - new Date(b.date);
        case "amount_desc":
          return b.amount - a.amount;
        case "amount_asc":
          return a.amount - b.amount;
        default:
          return new Date(b.date) - new Date(a.date);
      }
    });
    if (pageItems.length === 0) {
      tbody.innerHTML = `
<tr>
<td colspan="7" class="px-6 py-12 text-center text-gray-500 dark:text-gray-400">
<i class="fas fa-inbox text-4xl mb-3 opacity-50"></i>
<p class="text-lg font-medium">Записи не найдены</p>
<p class="text-sm mt-1">Попробуйте изменить фильтры или добавить новую запись</p>
</td>
</tr>
`;
      this.updatePagination();
      return;
    }
    tbody.innerHTML = pageItems
      .map(
        (item) => `
<tr class="hover:bg-gray-50 dark:hover:bg-gray-700/50 theme-transition">
<td class="px-6 py-4 whitespace-nowrap text-sm font-medium">${item.id}</td>
<td class="px-6 py-4 whitespace-nowrap">
<span class="px-3 py-1 rounded-full text-xs font-medium ${
          item.type === "income"
            ? "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300"
            : "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300"
        }">
${item.type === "income" ? '<i class="fas fa-arrow-up mr-1"></i>Доход' : '<i class="fas fa-arrow-down mr-1"></i>Расход'}
</span>
</td>
<td class="px-6 py-4 whitespace-nowrap text-sm font-bold ${
          item.type === "income"
            ? "text-green-600 dark:text-green-400"
            : "text-red-600 dark:text-red-400"
        }">
${item.amount.toLocaleString("ru-RU", { minimumFractionDigits: 2, maximumFractionDigits: 2 })} ₽
</td>
<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
${new Date(item.date).toLocaleDateString("ru-RU", {
  day: "2-digit",
  month: "short",
  year: "numeric",
  hour: "2-digit",
  minute: "2-digit",
})}
</td>
<td class="px-6 py-4 whitespace-nowrap">
<span class="px-2 py-1 bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300 rounded text-xs font-medium">
${item.category || "Без категории"}
</span>
</td>
<td class="px-6 py-4 text-sm text-gray-500 dark:text-gray-400 truncate max-w-xs" title="${item.description || ""}">
${item.description || "-"}
</td>
<td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
<div class="flex space-x-2">
<button onclick="app.openEditModal(${item.id})" class="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300">
<i class="fas fa-edit"></i>
</button>
<button onclick="app.confirmDelete(${item.id})" class="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300">
<i class="fas fa-trash-alt"></i>
</button>
</div>
</td>
</tr>
`,
      )
      .join("");
    this.updatePagination();
  }
  updatePagination() {
    document.getElementById("currentPageStart").textContent = Math.min(
      (this.currentPage - 1) * this.itemsPerPage + 1,
      this.totalItems,
    );
    document.getElementById("currentPageEnd").textContent = Math.min(
      this.currentPage * this.itemsPerPage,
      this.totalItems,
    );
    document.getElementById("totalItems").textContent = this.totalItems;
    const prevBtn = document.getElementById("prevPage");
    const nextBtn = document.getElementById("nextPage");
    prevBtn.disabled = this.currentPage === 1;
    nextBtn.disabled = this.currentPage * this.itemsPerPage >= this.totalItems;
    prevBtn.classList.toggle("opacity-50", prevBtn.disabled);
    prevBtn.classList.toggle("cursor-not-allowed", prevBtn.disabled);
    nextBtn.classList.toggle("opacity-50", nextBtn.disabled);
    nextBtn.classList.toggle("cursor-not-allowed", nextBtn.disabled);
  }
  changePage(delta) {
    this.currentPage += delta;
    if (this.currentPage < 1) this.currentPage = 1;
    if ((this.currentPage - 1) * this.itemsPerPage >= this.totalItems)
      this.currentPage = Math.ceil(this.totalItems / this.itemsPerPage) || 1;
    this.loadItems();
    window.scrollTo({ top: 0, behavior: "smooth" });
  }
  switchTab(tabName) {
    document
      .querySelectorAll(".tab-content")
      .forEach((tab) => tab.classList.add("hidden"));
    document
      .querySelectorAll(".tab-btn")
      .forEach((btn) =>
        btn.classList.remove("border-primary", "text-white", "font-bold"),
      );
    document.getElementById(`${tabName}Tab`).classList.remove("hidden");
    document
      .querySelector(`[data-tab="${tabName}"]`)
      .classList.add("border-primary", "text-white", "font-bold");
    this.currentTab = tabName;
    if (tabName === "analytics" && !this.analyticsLoaded) {
      this.loadAnalyticsPresets();
    }
  }
  async handleAddItem(e) {
    e.preventDefault();
    const form = e.target;
    const amount = parseFloat(document.getElementById("itemAmount").value);
    if (isNaN(amount) || amount <= 0) {
      this.showError(
        "Некорректная сумма",
        "Сумма должна быть положительным числом",
      );
      return;
    }
    const newItem = {
      type: document.getElementById("itemType").value,
      amount: amount,
      date: document.getElementById("itemDate").value,
      category:
        document.getElementById("itemCategory").value.trim() || "Без категории",
      description: document.getElementById("itemDescription").value.trim(),
    };
    try {
      const response = await fetch("/items", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(newItem),
      });
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || "Ошибка создания записи");
      }
      const result = await response.json();
      this.showSuccess("Запись добавлена", `ID: ${result.id}`);
      form.reset();
      const now = new Date();
      document.getElementById("itemDate").valueAsNumber =
        now.getTime() - now.getTimezoneOffset() * 60000;
      this.loadItems();
    } catch (error) {
      this.showError("Ошибка добавления записи", error.message);
      console.error("Add item error:", error);
    }
  }
  async openEditModal(id) {
    try {
      const response = await fetch(`/items/${id}`);
      if (!response.ok) throw new Error("Запись не найдена");
      const item = await response.json();
      document.getElementById("editItemId").value = item.id;
      document.getElementById("editItemType").value = item.type;
      document.getElementById("editItemAmount").value = item.amount;
      const date = new Date(item.date);
      const localDate = new Date(
        date.getTime() - date.getTimezoneOffset() * 60000,
      );
      document.getElementById("editItemDate").value = localDate
        .toISOString()
        .slice(0, 16);
      document.getElementById("editItemCategory").value = item.category || "";
      document.getElementById("editItemDescription").value =
        item.description || "";
      document.getElementById("editModal").classList.remove("hidden");
      document.body.style.overflow = "hidden";
    } catch (error) {
      this.showError("Ошибка загрузки записи", error.message);
    }
  }
  closeEditModal() {
    document.getElementById("editModal").classList.add("hidden");
    document.body.style.overflow = "";
  }
  async handleUpdateItem(e) {
    e.preventDefault();
    const id = document.getElementById("editItemId").value;
    const updatedItem = {
      type: document.getElementById("editItemType").value,
      amount: parseFloat(document.getElementById("editItemAmount").value),
      date: document.getElementById("editItemDate").value,
      category:
        document.getElementById("editItemCategory").value.trim() ||
        "Без категории",
      description: document.getElementById("editItemDescription").value.trim(),
    };
    try {
      const response = await fetch(`/items/${id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(updatedItem),
      });
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || "Ошибка обновления записи");
      }
      this.showSuccess("Запись обновлена", `ID: ${id}`);
      this.closeEditModal();
      this.loadItems();
    } catch (error) {
      this.showError("Ошибка обновления записи", error.message);
    }
  }
  async confirmDelete(id) {
    if (
      !confirm(
        "Вы уверены, что хотите удалить эту запись? Это действие нельзя отменить.",
      )
    )
      return;
    await this.deleteItem(id);
  }
  async confirmDeleteItem() {
    const id = document.getElementById("editItemId").value;
    if (
      !confirm(
        "Вы уверены, что хотите удалить эту запись? Это действие нельзя отменить.",
      )
    )
      return;
    await this.deleteItem(id);
    this.closeEditModal();
  }
  async deleteItem(id) {
    try {
      const response = await fetch(`/items/${id}`, { method: "DELETE" });
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || "Ошибка удаления записи");
      }
      this.showSuccess("Запись удалена", `ID: ${id}`);
      this.loadItems();
    } catch (error) {
      this.showError("Ошибка удаления записи", error.message);
    }
  }
  applyFilters() {
    const typeFilter = document.getElementById("filterType").value;
    const categoryFilter = document.getElementById("filterCategory").value;
    const periodFilter = document.getElementById("filterPeriod").value;
    const searchTerm = document
      .getElementById("globalSearch")
      .value.toLowerCase();
    let filtered = [...this.itemsCache];
    if (typeFilter) {
      filtered = filtered.filter((item) => item.type === typeFilter);
    }
    if (categoryFilter) {
      filtered = filtered.filter((item) => item.category === categoryFilter);
    }
    if (periodFilter !== "all") {
      const now = new Date();
      let startDate = new Date();
      switch (periodFilter) {
        case "today":
          startDate = new Date(
            now.getFullYear(),
            now.getMonth(),
            now.getDate(),
          );
          break;
        case "week":
          startDate = new Date(now.setDate(now.getDate() - 7));
          break;
        case "month":
          startDate = new Date(now.setMonth(now.getMonth() - 1));
          break;
        case "quarter":
          startDate = new Date(now.setMonth(now.getMonth() - 3));
          break;
      }
      filtered = filtered.filter((item) => new Date(item.date) >= startDate);
    }
    if (searchTerm) {
      filtered = filtered.filter(
        (item) =>
          item.category?.toLowerCase().includes(searchTerm) ||
          item.description?.toLowerCase().includes(searchTerm) ||
          item.amount.toString().includes(searchTerm) ||
          item.type.includes(searchTerm),
      );
    }
    this.itemsCache = filtered;
    this.totalItems = filtered.length;
    this.currentPage = 1;
    this.renderItemsTable();
    this.showInfo("Фильтры применены", `${this.totalItems} записей`);
  }
  globalSearch(term) {
    if (term.length > 2 || term.length === 0) {
      this.applyFilters();
    }
  }
  async handleGetAnalytics(e) {
    e.preventDefault();
    const from = document.getElementById("analyticsFrom").value;
    const to = document.getElementById("analyticsTo").value;
    if (!from || !to) {
      this.showError("Ошибка периода", "Выберите начальную и конечную дату");
      return;
    }
    if (new Date(from) > new Date(to)) {
      this.showError(
        "Ошибка периода",
        "Начальная дата не может быть позже конечной",
      );
      return;
    }
    try {
      document.getElementById("analyticsResult").classList.add("hidden");
      document
        .querySelector("#analyticsTab .fa-chart-bar")
        .classList.add("fa-spin");
      const response = await fetch(
        `/analytics?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`,
      );
      document
        .querySelector("#analyticsTab .fa-chart-bar")
        .classList.remove("fa-spin");
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || "Ошибка получения аналитики");
      }
      const analytics = await response.json();
      this.renderAnalytics(analytics, from, to);
      this.showSuccess(
        "Аналитика получена",
        `Период: ${new Date(from).toLocaleDateString()} - ${new Date(to).toLocaleDateString()}`,
      );
    } catch (error) {
      document
        .querySelector("#analyticsTab .fa-chart-bar")
        .classList.remove("fa-spin");
      this.showError("Ошибка аналитики", error.message);
      console.error("Analytics error:", error);
    }
  }
  renderAnalytics(analytics, from, to) {
    document.getElementById("sumResult").textContent =
      analytics.sum.toLocaleString("ru-RU", {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      }) + " ₽";
    document.getElementById("avgResult").textContent =
      analytics.avg.toLocaleString("ru-RU", {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      }) + " ₽";
    document.getElementById("countResult").textContent =
      analytics.count.toLocaleString("ru-RU");
    document.getElementById("medianResult").textContent =
      analytics.median.toLocaleString("ru-RU", {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      }) + " ₽";
    document.getElementById("percent90Result").textContent =
      analytics.percent90.toLocaleString("ru-RU", {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      }) + " ₽";
    this.renderAnalyticsDetails(analytics.details || [], from, to);
    if (document.getElementById("includeCharts").checked) {
      this.updateCharts(analytics.details || []);
    }
    document.getElementById("analyticsResult").classList.remove("hidden");
    this.analyticsLoaded = true;
  }
  renderAnalyticsDetails(details, from, to) {
    const tbody = document.getElementById("analyticsDetails");
    if (!details || details.length === 0) {
      tbody.innerHTML = `
<tr>
<td colspan="4" class="px-6 py-12 text-center text-gray-500 dark:text-gray-400">
<i class="fas fa-chart-bar text-4xl mb-3 opacity-50"></i>
<p class="text-lg font-medium">Нет данных за выбранный период</p>
<p class="text-sm mt-1">Попробуйте выбрать другой период или добавить записи</p>
</td>
</tr>
`;
      return;
    }
    tbody.innerHTML = details
      .map(
        (item) => `
<tr class="hover:bg-gray-50 dark:hover:bg-gray-700/50 theme-transition">
<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
${new Date(item.date).toLocaleDateString("ru-RU", {
  day: "2-digit",
  month: "short",
  year: "numeric",
  hour: "2-digit",
  minute: "2-digit",
})}
</td>
<td class="px-6 py-4 whitespace-nowrap">
<span class="px-3 py-1 rounded-full text-xs font-medium ${
          item.type === "income"
            ? "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300"
            : "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300"
        }">
${item.type === "income" ? "Доход" : "Расход"}
</span>
</td>
<td class="px-6 py-4 whitespace-nowrap text-sm font-bold ${
          item.type === "income"
            ? "text-green-600 dark:text-green-400"
            : "text-red-600 dark:text-red-400"
        }">
${item.amount.toLocaleString("ru-RU", { minimumFractionDigits: 2, maximumFractionDigits: 2 })} ₽
</td>
<td class="px-6 py-4 whitespace-nowrap">
<span class="px-2 py-1 bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300 rounded text-xs font-medium">
${item.category || "Без категории"}
</span>
</td>
</tr>
`,
      )
      .join("");
  }
  setupCharts() {
    this.categoryChart = new Chart(document.getElementById("categoryChart"), {
      type: "doughnut",
      data: {
        labels: [],
        datasets: [
          {
            data: [],
            backgroundColor: [
              "#4361ee",
              "#3f37c9",
              "#4895ef",
              "#4cc9f0",
              "#7209b7",
              "#f72585",
              "#b5179e",
              "#7209b7",
            ],
            borderWidth: 0,
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            position: "bottom",
            labels: {
              color: document.documentElement.classList.contains("dark")
                ? "#e2e8f0"
                : "#4b5563",
              padding: 20,
              font: {
                size: 12,
              },
            },
          },
          tooltip: {
            callbacks: {
              label: (context) => {
                const label = context.label || "";
                const value = context.raw || 0;
                const total = context.dataset.data.reduce((a, b) => a + b, 0);
                const percentage = Math.round((value / total) * 100);
                return `${label}: ${value.toLocaleString("ru-RU")} ₽ (${percentage}%)`;
              },
            },
          },
        },
        cutout: "60%",
      },
    });
    this.timelineChart = new Chart(document.getElementById("timelineChart"), {
      type: "bar",
      data: {
        labels: [],
        datasets: [
          {
            label: "Доход",
            data: [],
            backgroundColor: "rgba(67, 97, 238, 0.7)",
            borderColor: "rgba(67, 97, 238, 1)",
            borderWidth: 1,
          },
          {
            label: "Расход",
            data: [],
            backgroundColor: "rgba(247, 37, 133, 0.7)",
            borderColor: "rgba(247, 37, 133, 1)",
            borderWidth: 1,
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            position: "top",
            labels: {
              color: document.documentElement.classList.contains("dark")
                ? "#e2e8f0"
                : "#4b5563",
              font: {
                size: 12,
              },
            },
          },
          tooltip: {
            callbacks: {
              label: (context) => {
                return `${context.dataset.label}: ${context.raw.toLocaleString("ru-RU")} ₽`;
              },
            },
          },
        },
        scales: {
          x: {
            ticks: {
              color: document.documentElement.classList.contains("dark")
                ? "#94a3b8"
                : "#6b7280",
            },
            grid: {
              color: document.documentElement.classList.contains("dark")
                ? "rgba(148, 163, 184, 0.1)"
                : "rgba(107, 114, 128, 0.1)",
            },
          },
          y: {
            ticks: {
              color: document.documentElement.classList.contains("dark")
                ? "#94a3b8"
                : "#6b7280",
              callback: (value) => `${value.toLocaleString("ru-RU")} ₽`,
            },
            grid: {
              color: document.documentElement.classList.contains("dark")
                ? "rgba(148, 163, 184, 0.1)"
                : "rgba(107, 114, 128, 0.1)",
            },
          },
        },
      },
    });
  }
  updateCharts(details) {
    const categoryData = {};
    details.forEach((item) => {
      const cat = item.category || "Без категории";
      if (!categoryData[cat]) categoryData[cat] = 0;
      categoryData[cat] += item.amount;
    });
    this.categoryChart.data.labels = Object.keys(categoryData);
    this.categoryChart.data.datasets[0].data = Object.values(categoryData);
    this.categoryChart.update();
    const dateGroups = {};
    details.forEach((item) => {
      const date = new Date(item.date).toLocaleDateString("ru-RU", {
        day: "2-digit",
        month: "short",
      });
      if (!dateGroups[date]) {
        dateGroups[date] = { income: 0, expense: 0 };
      }
      if (item.type === "income") {
        dateGroups[date].income += item.amount;
      } else {
        dateGroups[date].expense += item.amount;
      }
    });
    const dates = Object.keys(dateGroups);
    this.timelineChart.data.labels = dates;
    this.timelineChart.data.datasets[0].data = dates.map(
      (date) => dateGroups[date].income,
    );
    this.timelineChart.data.datasets[1].data = dates.map(
      (date) => dateGroups[date].expense,
    );
    this.timelineChart.update();
  }
  exportAllToCSV() {
    if (this.itemsCache.length === 0) {
      this.showError("Экспорт невозможен", "Нет записей для экспорта");
      return;
    }
    const headers = ["ID", "Тип", "Сумма (₽)", "Дата", "Категория", "Описание"];
    const rows = this.itemsCache.map((item) => [
      item.id,
      item.type === "income" ? "Доход" : "Расход",
      item.amount.toFixed(2),
      new Date(item.date).toLocaleString("ru-RU"),
      item.category || "",
      `"${(item.description || "").replace(/"/g, '""')}"`,
    ]);
    let csvContent =
      "data:text/csv;charset=utf-8,%EF%BB%BF" +
      headers.join(",") +
      "\n" +
      rows.map((row) => row.join(",")).join("\n");
    const encodedUri = encodeURI(csvContent);
    const link = document.createElement("a");
    link.setAttribute("href", encodedUri);
    link.setAttribute(
      "download",
      `sales_tracker_export_${new Date().toISOString().slice(0, 10)}.csv`,
    );
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    this.showSuccess(
      "Экспорт завершён",
      `Экспортировано ${this.itemsCache.length} записей`,
    );
  }
  exportAnalyticsToCSV() {
    const details = document.querySelectorAll("#analyticsDetails tr");
    if (
      details.length === 0 ||
      (details.length === 1 && details[0].querySelector("td"))
    ) {
      this.showError("Экспорт невозможен", "Нет данных аналитики для экспорта");
      return;
    }
    const headers = ["Дата", "Тип", "Сумма (₽)", "Категория"];
    const rows = Array.from(details).map((row) => {
      const cells = row.querySelectorAll("td");
      return [
        cells[0].textContent.trim(),
        cells[1].querySelector("span").textContent.trim(),
        cells[2].textContent.trim().replace(" ₽", ""),
        cells[3].textContent.trim(),
      ];
    });
    let csvContent =
      "data:text/csv;charset=utf-8,%EF%BB%BF" +
      headers.join(",") +
      "\n" +
      rows.map((row) => row.join(",")).join("\n");
    const encodedUri = encodeURI(csvContent);
    const link = document.createElement("a");
    link.setAttribute("href", encodedUri);
    link.setAttribute(
      "download",
      `analytics_export_${new Date().toISOString().slice(0, 10)}.csv`,
    );
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    this.showSuccess("Экспорт аналитики", "Данные сохранены в CSV");
  }
  generateReport() {
    const reportType = document.getElementById("reportType").value;
    const exportFormat = document.getElementById("exportFormat").value;
    const from = document.getElementById("reportFrom").value;
    const to = document.getElementById("reportTo").value;
    if (!from || !to) {
      this.showError("Ошибка периода", "Выберите период отчёта");
      return;
    }
    setTimeout(() => {
      this.showSuccess(
        "Отчёт сгенерирован",
        `${reportType} за период (${exportFormat.toUpperCase()})`,
      );
      const historyRow = document.createElement("tr");
      historyRow.innerHTML = `
<td class="px-6 py-4 whitespace-nowrap text-sm">${new Date().toLocaleString("ru-RU")}</td>
<td class="px-6 py-4 whitespace-nowrap">${this.getReportTypeName(reportType)}</td>
<td class="px-6 py-4 whitespace-nowrap">${new Date(from).toLocaleDateString()} - ${new Date(to).toLocaleDateString()}</td>
<td class="px-6 py-4 whitespace-nowrap">${exportFormat.toUpperCase()}</td>
<td class="px-6 py-4 whitespace-nowrap">
<button class="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300">
<i class="fas fa-download mr-1"></i>Скачать
</button>
</td>
`;
      const historyTable = document.getElementById("reportsHistory");
      if (historyTable.querySelector("tr td[colspan]")) {
        historyTable.innerHTML = "";
      }
      historyTable.insertBefore(historyRow, historyTable.firstChild);
      document
        .getElementById("reportsTab")
        .scrollIntoView({ behavior: "smooth" });
    }, 800);
  }
  getReportTypeName(type) {
    const types = {
      summary: "Сводный",
      category: "По категориям",
      period: "По периодам",
      custom: "Пользовательский",
    };
    return types[type] || type;
  }
  loadAnalyticsPresets() {
    const now = new Date();
    const todayStart = new Date(
      now.getFullYear(),
      now.getMonth(),
      now.getDate(),
      0,
      0,
      0,
    );
    const todayEnd = new Date(
      now.getFullYear(),
      now.getMonth(),
      now.getDate(),
      23,
      59,
      59,
    );
    document.getElementById("analyticsFrom").valueAsNumber =
      todayStart.getTime() - todayStart.getTimezoneOffset() * 60000;
    document.getElementById("analyticsTo").valueAsNumber =
      todayEnd.getTime() - todayEnd.getTimezoneOffset() * 60000;
  }
  showToast(message, type = "success", details = "") {
    const toast = document.getElementById("toast");
    const icon = document.getElementById("toastIcon");
    const messageEl = document.getElementById("toastMessage");
    const detailsEl = document.getElementById("toastDetails");
    toast.className =
      "fixed bottom-6 right-6 max-w-sm bg-white dark:bg-gray-800 border-l-4 text-gray-800 dark:text-gray-200 p-4 rounded-lg shadow-lg transform transition-transform duration-300 z-50 theme-transition";
    switch (type) {
      case "success":
        toast.classList.add("border-green-500");
        icon.className = "fas fa-check-circle text-green-500 text-xl";
        break;
      case "error":
        toast.classList.add("border-red-500");
        icon.className = "fas fa-exclamation-circle text-red-500 text-xl";
        break;
      case "info":
        toast.classList.add("border-blue-500");
        icon.className = "fas fa-info-circle text-blue-500 text-xl";
        break;
      case "warning":
        toast.classList.add("border-yellow-500");
        icon.className = "fas fa-exclamation-triangle text-yellow-500 text-xl";
        break;
    }
    messageEl.textContent = message;
    detailsEl.textContent = details;
    detailsEl.classList.toggle("hidden", !details);
    toast.classList.remove("translate-x-full");
    toast.classList.add("translate-x-0");
    setTimeout(() => this.hideToast(), 5000);
  }
  hideToast() {
    const toast = document.getElementById("toast");
    toast.classList.remove("translate-x-0");
    toast.classList.add("translate-x-full");
  }
  showSuccess(message, details = "") {
    this.showToast(message, "success", details);
  }
  showError(message, details = "") {
    this.showToast(message, "error", details);
  }
  showInfo(message, details = "") {
    this.showToast(message, "info", details);
  }
  showWarning(message, details = "") {
    this.showToast(message, "warning", details);
  }
}
document.addEventListener("DOMContentLoaded", () => {
  window.app = new SalesTrackerApp();
  document.addEventListener("keydown", (e) => {
    if (e.ctrlKey && e.key === "n") {
      e.preventDefault();
      document.getElementById("itemAmount").focus();
    }
  });
});
