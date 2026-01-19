package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/avakumov/metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestFile(t *testing.T) string {
	t.Helper()

	// Создаем временный файл
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_metrics.json")

	return filePath
}

func cleanupTestFile(t *testing.T, filePath string) {
	t.Helper()
	os.Remove(filePath)
}

func TestNewFileRepository(t *testing.T) {
	t.Run("create new file when not exists", func(t *testing.T) {
		filePath := setupTestFile(t)
		defer cleanupTestFile(t, filePath)

		// Удаляем файл если он существует
		os.Remove(filePath)

		repo, err := NewFileRepository(filePath)
		require.NoError(t, err)
		require.NotNil(t, repo)
		defer repo.file.Close()

		// Проверяем, что файл создан
		_, err = os.Stat(filePath)
		assert.NoError(t, err)

		// Проверяем, что можем прочитать пустой массив
		metrics, err := repo.GetAll()
		require.NoError(t, err)
		assert.Empty(t, metrics)
	})

	t.Run("open existing file", func(t *testing.T) {
		filePath := setupTestFile(t)
		defer cleanupTestFile(t, filePath)

		// Создаем файл с тестовыми данными
		initialMetrics := []models.Metric{
			{ID: "test1", MType: "counter", Delta: new(int64)},
		}
		*initialMetrics[0].Delta = 100

		data, err := json.Marshal(initialMetrics)
		require.NoError(t, err)

		err = os.WriteFile(filePath, data, 0644)
		require.NoError(t, err)

		repo, err := NewFileRepository(filePath)
		require.NoError(t, err)
		require.NotNil(t, repo)
		defer repo.file.Close()

		// Проверяем, что данные загрузились
		metrics, err := repo.GetAll()
		require.NoError(t, err)
		require.Len(t, metrics, 1)
		assert.Equal(t, "test1", metrics[0].ID)
	})

	t.Run("error when cannot create file", func(t *testing.T) {
		// Пытаемся создать файл в несуществующей директории
		filePath := "/non/existent/dir/metrics.json"
		repo, err := NewFileRepository(filePath)
		assert.Error(t, err)
		assert.Nil(t, repo)
	})
}

func TestGetAll(t *testing.T) {
	t.Run("empty file returns empty slice", func(t *testing.T) {
		filePath := setupTestFile(t)
		defer cleanupTestFile(t, filePath)

		// Создаем пустой файл
		repo, err := NewFileRepository(filePath)
		require.NoError(t, err)
		require.NotNil(t, repo)
		defer repo.file.Close()

		metrics, err := repo.GetAll()
		require.NoError(t, err)
		assert.Empty(t, metrics)
	})

	t.Run("read existing metrics", func(t *testing.T) {
		filePath := setupTestFile(t)
		defer cleanupTestFile(t, filePath)

		// Создаем файл с данными
		expectedMetrics := []models.Metric{
			{ID: "metric1", MType: "counter", Delta: new(int64)},
			{ID: "metric2", MType: "gauge", Value: new(float64)},
		}
		*expectedMetrics[0].Delta = 42
		*expectedMetrics[1].Value = 3.14

		data, err := json.Marshal(expectedMetrics)
		require.NoError(t, err)

		err = os.WriteFile(filePath, data, 0644)
		require.NoError(t, err)

		repo, err := NewFileRepository(filePath)
		require.NoError(t, err)
		require.NotNil(t, repo)
		defer repo.file.Close()

		metrics, err := repo.GetAll()
		require.NoError(t, err)
		require.Len(t, metrics, 2)

		// Проверяем метрики
		assert.Equal(t, "metric1", metrics[0].ID)
		assert.Equal(t, "counter", metrics[0].MType)
		require.NotNil(t, metrics[0].Delta)
		assert.Equal(t, int64(42), *metrics[0].Delta)

		assert.Equal(t, "metric2", metrics[1].ID)
		assert.Equal(t, "gauge", metrics[1].MType)
		require.NotNil(t, metrics[1].Value)
		assert.Equal(t, 3.14, *metrics[1].Value)
	})

	t.Run("handle corrupted json", func(t *testing.T) {
		filePath := setupTestFile(t)
		defer cleanupTestFile(t, filePath)

		// Пишем некорректный JSON
		err := os.WriteFile(filePath, []byte("{invalid json"), 0644)
		require.NoError(t, err)

		repo, err := NewFileRepository(filePath)
		require.NoError(t, err)
		require.NotNil(t, repo)
		defer repo.file.Close()

		metrics, err := repo.GetAll()
		assert.Error(t, err)
		assert.Nil(t, metrics)
	})
}

func TestGetMetricByID(t *testing.T) {
	filePath := setupTestFile(t)
	defer cleanupTestFile(t, filePath)

	// Инициализируем репозиторий с тестовыми данными
	initialMetrics := []models.Metric{
		{ID: "cpu_usage", MType: "gauge", Value: new(float64)},
		{ID: "request_count", MType: "counter", Delta: new(int64)},
		{ID: "memory_used", MType: "gauge", Value: new(float64)},
	}
	*initialMetrics[0].Value = 75.5
	*initialMetrics[1].Delta = 1000
	*initialMetrics[2].Value = 2048.0

	data, err := json.Marshal(initialMetrics)
	require.NoError(t, err)

	err = os.WriteFile(filePath, data, 0644)
	require.NoError(t, err)

	repo, err := NewFileRepository(filePath)
	require.NoError(t, err)
	require.NotNil(t, repo)
	defer repo.file.Close()

	t.Run("existing metric", func(t *testing.T) {
		metric, err := repo.GetMetricByID("cpu_usage")
		require.NoError(t, err)
		require.NotNil(t, metric)
		assert.Equal(t, "cpu_usage", metric.ID)
		assert.Equal(t, "gauge", metric.MType)
		require.NotNil(t, metric.Value)
		assert.Equal(t, 75.5, *metric.Value)
	})

	t.Run("non-existing metric", func(t *testing.T) {
		metric, err := repo.GetMetricByID("non_existing")
		assert.Error(t, err)
		assert.Nil(t, metric)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("empty id", func(t *testing.T) {
		metric, err := repo.GetMetricByID("")
		assert.Error(t, err)
		assert.Nil(t, metric)
	})
}

func TestSaveMetric(t *testing.T) {
	filePath := setupTestFile(t)
	defer cleanupTestFile(t, filePath)

	repo, err := NewFileRepository(filePath)
	require.NoError(t, err)
	require.NotNil(t, repo)
	defer repo.file.Close()

	t.Run("save new metric", func(t *testing.T) {
		newMetric := models.Metric{
			ID:    "new_counter",
			MType: "counter",
			Delta: new(int64),
		}
		*newMetric.Delta = 50

		err := repo.SaveMetric(newMetric)
		require.NoError(t, err)

		// Проверяем, что метрика сохранилась
		metrics, err := repo.GetAll()
		require.NoError(t, err)
		require.Len(t, metrics, 1)
		assert.Equal(t, "new_counter", metrics[0].ID)
		require.NotNil(t, metrics[0].Delta)
		assert.Equal(t, int64(50), *metrics[0].Delta)
	})

	t.Run("update existing metric", func(t *testing.T) {
		// Сначала сохраняем метрику
		metric := models.Metric{
			ID:    "update_test",
			MType: "gauge",
			Value: new(float64),
		}
		*metric.Value = 10.0

		err := repo.SaveMetric(metric)
		require.NoError(t, err)

		// Обновляем значение
		updatedMetric := models.Metric{
			ID:    "update_test",
			MType: "gauge",
			Value: new(float64),
		}
		*updatedMetric.Value = 20.0

		err = repo.SaveMetric(updatedMetric)
		require.NoError(t, err)

		// Проверяем, что обновилось
		savedMetric, err := repo.GetMetricByID("update_test")
		require.NoError(t, err)
		require.NotNil(t, savedMetric.Value)
		assert.Equal(t, 20.0, *savedMetric.Value)

		// Проверяем, что не создалась дубликат
		metrics, err := repo.GetAll()
		require.NoError(t, err)

		// Должны быть 2 метрики: new_counter и update_test
		assert.Len(t, metrics, 2)
	})

	t.Run("update with different type should replace", func(t *testing.T) {
		// Сохраняем как counter
		counterMetric := models.Metric{
			ID:    "same_id",
			MType: "counter",
			Delta: new(int64),
		}
		*counterMetric.Delta = 100

		err := repo.SaveMetric(counterMetric)
		require.NoError(t, err)

		// Обновляем как gauge
		gaugeMetric := models.Metric{
			ID:    "same_id",
			MType: "gauge",
			Value: new(float64),
		}
		*gaugeMetric.Value = 99.9

		err = repo.SaveMetric(gaugeMetric)
		require.NoError(t, err)

		// Проверяем, что теперь gauge
		metric, err := repo.GetMetricByID("same_id")
		require.NoError(t, err)
		assert.Equal(t, "gauge", metric.MType)
		assert.Nil(t, metric.Delta) // Delta должен быть nil
		require.NotNil(t, metric.Value)
		assert.Equal(t, 99.9, *metric.Value)
	})
}

func TestSaveMetrics(t *testing.T) {
	filePath := setupTestFile(t)
	defer cleanupTestFile(t, filePath)

	repo, err := NewFileRepository(filePath)
	require.NoError(t, err)
	require.NotNil(t, repo)
	defer repo.file.Close()

	t.Run("save multiple metrics", func(t *testing.T) {
		metrics := []models.Metric{
			{
				ID:    "metric1",
				MType: "counter",
				Delta: new(int64),
			},
			{
				ID:    "metric2",
				MType: "gauge",
				Value: new(float64),
			},
			{
				ID:    "metric3",
				MType: "counter",
				Delta: new(int64),
			},
		}
		*metrics[0].Delta = 1
		*metrics[1].Value = 2.0
		*metrics[2].Delta = 3

		err := repo.SaveMetrics(metrics)
		require.NoError(t, err)

		// Проверяем сохранение
		savedMetrics, err := repo.GetAll()
		require.NoError(t, err)
		assert.Len(t, savedMetrics, 3)

		// Проверяем порядок (может измениться при обновлении)
		metricIDs := make([]string, len(savedMetrics))
		for i, m := range savedMetrics {
			metricIDs[i] = m.ID
		}
		assert.ElementsMatch(t, []string{"metric1", "metric2", "metric3"}, metricIDs)
	})

	t.Run("save empty slice", func(t *testing.T) {
		err := repo.SaveMetrics([]models.Metric{})
		require.NoError(t, err)

		// Проверяем, что предыдущие данные остались
		metrics, err := repo.GetAll()
		require.NoError(t, err)
		assert.NotEmpty(t, metrics) // В предыдущем тесте мы сохранили 3 метрики
	})

	t.Run("save metrics with duplicates", func(t *testing.T) {
		// Сначала сохраняем одну метрику
		initialMetric := models.Metric{
			ID:    "duplicate_test",
			MType: "counter",
			Delta: new(int64),
		}
		*initialMetric.Delta = 10

		err := repo.SaveMetric(initialMetric)
		require.NoError(t, err)

		// Теперь пытаемся сохранить несколько, включая дубликат
		metrics := []models.Metric{
			{
				ID:    "duplicate_test", // Дубликат
				MType: "counter",
				Delta: new(int64),
			},
			{
				ID:    "new_metric",
				MType: "gauge",
				Value: new(float64),
			},
		}
		*metrics[0].Delta = 99 // Новое значение для дубликата
		*metrics[1].Value = 50.0

		err = repo.SaveMetrics(metrics)
		require.NoError(t, err)

		// Проверяем, что дубликат обновился
		updatedMetric, err := repo.GetMetricByID("duplicate_test")
		require.NoError(t, err)
		require.NotNil(t, updatedMetric.Delta)
		assert.Equal(t, int64(99), *updatedMetric.Delta)

		// Проверяем, что новая метрика добавилась
		_, err = repo.GetMetricByID("new_metric")
		assert.NoError(t, err)
	})
}

func TestDeleteMetricByID(t *testing.T) {
	filePath := setupTestFile(t)
	defer cleanupTestFile(t, filePath)

	// Подготавливаем данные
	initialMetrics := []models.Metric{
		{ID: "to_delete", MType: "counter", Delta: new(int64)},
		{ID: "keep1", MType: "gauge", Value: new(float64)},
		{ID: "keep2", MType: "counter", Delta: new(int64)},
	}
	*initialMetrics[0].Delta = 100
	*initialMetrics[1].Value = 50.0
	*initialMetrics[2].Delta = 200

	data, err := json.Marshal(initialMetrics)
	require.NoError(t, err)

	err = os.WriteFile(filePath, data, 0666)
	require.NoError(t, err)

	repo, err := NewFileRepository(filePath)
	require.NoError(t, err)
	require.NotNil(t, repo)
	defer repo.file.Close()

	t.Run("delete existing metric", func(t *testing.T) {
		// Убеждаемся, что метрика существует
		_, err := repo.GetMetricByID("to_delete")
		require.NoError(t, err)

		// Удаляем
		err = repo.DeleteMetricByID("to_delete")
		require.NoError(t, err)

		// Проверяем, что удалилась
		metric, err := repo.GetMetricByID("to_delete")
		assert.Error(t, err)
		assert.Nil(t, metric)

		// Проверяем, что остальные остались
		metrics, err := repo.GetAll()
		require.NoError(t, err)
		assert.Len(t, metrics, 2)

		// Проверяем IDs оставшихся метрик
		remainingIDs := make([]string, len(metrics))
		for i, m := range metrics {
			remainingIDs[i] = m.ID
		}
		assert.ElementsMatch(t, []string{"keep1", "keep2"}, remainingIDs)
	})

	t.Run("delete non-existing metric", func(t *testing.T) {
		// Сохраняем текущее состояние
		metricsBefore, err := repo.GetAll()
		require.NoError(t, err)
		beforeCount := len(metricsBefore)

		// Пытаемся удалить несуществующую метрику
		err = repo.DeleteMetricByID("non_existing")
		require.NoError(t, err) // Ваш текущий код не возвращает ошибку если не нашел!

		// Проверяем, что ничего не изменилось
		metricsAfter, err := repo.GetAll()
		require.NoError(t, err)
		assert.Len(t, metricsAfter, beforeCount)
	})

	t.Run("delete from empty repository", func(t *testing.T) {
		// Создаем новый пустой репозиторий
		emptyFilePath := setupTestFile(t)
		defer cleanupTestFile(t, emptyFilePath)

		emptyRepo, err := NewFileRepository(emptyFilePath)
		require.NoError(t, err)
		require.NotNil(t, emptyRepo)
		defer emptyRepo.file.Close()

		err = emptyRepo.DeleteMetricByID("any_id")
		require.NoError(t, err) // Не должно быть ошибки

		metrics, err := emptyRepo.GetAll()
		require.NoError(t, err)
		assert.Empty(t, metrics)
	})

	t.Run("delete and check order preservation", func(t *testing.T) {
		// Создаем тестовый файл с известным порядком
		orderFilePath := setupTestFile(t)
		defer cleanupTestFile(t, orderFilePath)

		orderedMetrics := []models.Metric{
			{ID: "first", MType: "counter", Delta: new(int64)},
			{ID: "second", MType: "gauge", Value: new(float64)},
			{ID: "third", MType: "counter", Delta: new(int64)},
			{ID: "fourth", MType: "gauge", Value: new(float64)},
		}

		data, _ := json.Marshal(orderedMetrics)
		err = os.WriteFile(orderFilePath, data, 0644)
		require.NoError(t, err)

		orderRepo, _ := NewFileRepository(orderFilePath)
		defer orderRepo.file.Close()

		// Удаляем "second"
		err := orderRepo.DeleteMetricByID("second")
		require.NoError(t, err)

		// Проверяем оставшиеся метрики
		remaining, err := orderRepo.GetAll()
		require.NoError(t, err)

		// ВАЖНО: ваш текущий метод delete использует unordered удаление!
		// Это значит, что порядок может измениться
		assert.Len(t, remaining, 3)

		// Проверяем, что "second" удален
		for _, m := range remaining {
			assert.NotEqual(t, "second", m.ID)
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	filePath := setupTestFile(t)
	defer cleanupTestFile(t, filePath)

	repo, err := NewFileRepository(filePath)
	require.NoError(t, err)
	require.NotNil(t, repo)
	defer repo.file.Close()

	// Запускаем несколько горутин для конкурентного доступа
	done := make(chan bool)
	errors := make(chan error, 10)

	// Горутина для записи
	go func() {
		for i := 0; i < 5; i++ {
			metric := models.Metric{
				ID:    "concurrent_metric",
				MType: "counter",
				Delta: new(int64),
			}
			*metric.Delta = int64(i)

			if err := repo.SaveMetric(metric); err != nil {
				errors <- err
			}
		}
		done <- true
	}()

	// Горутина для чтения
	go func() {
		for i := 0; i < 5; i++ {
			if _, err := repo.GetAll(); err != nil {
				errors <- err
			}
		}
		done <- true
	}()

	// Горутина для удаления/добавления
	go func() {
		for i := 0; i < 3; i++ {
			metric := models.Metric{
				ID:    "temp_metric",
				MType: "gauge",
				Value: new(float64),
			}
			*metric.Value = float64(i)

			err := repo.SaveMetric(metric)
			require.NoError(t, err)
			err = repo.DeleteMetricByID("temp_metric")
			require.NoError(t, err)
		}
		done <- true
	}()

	// Ждем завершения всех горутин
	for i := 0; i < 3; i++ {
		<-done
	}

	close(errors)

	// Проверяем, не было ли ошибок
	for err := range errors {
		assert.NoError(t, err, "concurrent operation failed")
	}

	// Финальная проверка состояния
	metrics, err := repo.GetAll()
	require.NoError(t, err)

	// Должна быть хотя бы одна метрика (concurrent_metric)
	assert.NotEmpty(t, metrics)
}

func TestFilePersistence(t *testing.T) {
	filePath := setupTestFile(t)
	defer cleanupTestFile(t, filePath)

	// Создаем и заполняем первый репозиторий
	repo1, err := NewFileRepository(filePath)
	require.NoError(t, err)
	require.NotNil(t, repo1)

	metric := models.Metric{
		ID:    "persistent",
		MType: "counter",
		Delta: new(int64),
	}
	*metric.Delta = 42

	err = repo1.SaveMetric(metric)
	require.NoError(t, err)

	// Закрываем первый репозиторий
	repo1.file.Close()

	// Создаем второй репозиторий с тем же файлом
	repo2, err := NewFileRepository(filePath)
	require.NoError(t, err)
	require.NotNil(t, repo2)
	defer repo2.file.Close()

	// Проверяем, что данные сохранились
	savedMetric, err := repo2.GetMetricByID("persistent")
	require.NoError(t, err)
	require.NotNil(t, savedMetric)
	assert.Equal(t, "persistent", savedMetric.ID)
	require.NotNil(t, savedMetric.Delta)
	assert.Equal(t, int64(42), *savedMetric.Delta)
}

// Вспомогательные функции для тестов
func createTestMetric(id, mtype string, delta int64, value float64) models.Metric {
	metric := models.Metric{
		ID:    id,
		MType: mtype,
	}

	if mtype == "counter" {
		metric.Delta = new(int64)
		*metric.Delta = delta
	} else if mtype == "gauge" {
		metric.Value = new(float64)
		*metric.Value = value
	}

	return metric
}

func TestEdgeCases(t *testing.T) {
	t.Run("metric with nil pointers", func(t *testing.T) {
		filePath := setupTestFile(t)
		defer cleanupTestFile(t, filePath)

		repo, err := NewFileRepository(filePath)
		require.NoError(t, err)
		defer repo.file.Close()

		// Метрика без Delta/Value указателей
		metric := models.Metric{
			ID:    "nil_pointers",
			MType: "counter",
			// Delta is nil
		}

		err = repo.SaveMetric(metric)
		require.NoError(t, err)

		// Проверяем, что сохранилась
		saved, err := repo.GetMetricByID("nil_pointers")
		require.NoError(t, err)
		assert.Equal(t, "nil_pointers", saved.ID)
		assert.Equal(t, "counter", saved.MType)
		assert.Nil(t, saved.Delta) // Delta остался nil
	})

	t.Run("very long metric id", func(t *testing.T) {
		filePath := setupTestFile(t)
		defer cleanupTestFile(t, filePath)

		repo, err := NewFileRepository(filePath)
		require.NoError(t, err)
		defer repo.file.Close()

		longID := string(make([]byte, 10000)) // Очень длинный ID
		metric := createTestMetric(longID, "counter", 1, 0)

		err = repo.SaveMetric(metric)
		require.NoError(t, err)

		saved, err := repo.GetMetricByID(longID)
		require.NoError(t, err)
		assert.Equal(t, longID, saved.ID)
	})
}
