package htmlViewer

import (
	"fmt"
	"html/template"
	"strings"
)

// GenerateIndexHTML generates an index.html file for local viewing of mosaic schema
func GenerateIndexHTML(couponCode string, pageCount int, stonesCount int) string {
	htmlTemplate := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Схема мозаики - {{.CouponCode}}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            align-items: center;
            padding: 20px;
        }
        
        .container {
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            max-width: 1200px;
            width: 100%;
            overflow: hidden;
        }
        
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        
        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
        }
        
        .header p {
            font-size: 1.2em;
            opacity: 0.9;
        }
        
        .info-panel {
            display: flex;
            justify-content: space-around;
            padding: 20px;
            background: #f8f9fa;
            border-bottom: 2px solid #e9ecef;
        }
        
        .info-item {
            text-align: center;
        }
        
        .info-item .label {
            color: #6c757d;
            font-size: 0.9em;
            margin-bottom: 5px;
        }
        
        .info-item .value {
            color: #495057;
            font-size: 1.8em;
            font-weight: bold;
        }
        
        .controls {
            padding: 30px;
            background: white;
            display: flex;
            flex-wrap: wrap;
            gap: 20px;
            justify-content: center;
            align-items: center;
        }
        
        .search-box {
            display: flex;
            gap: 10px;
            align-items: center;
        }
        
        .search-box label {
            font-weight: 500;
            color: #495057;
        }
        
        .search-box input {
            padding: 10px 15px;
            border: 2px solid #dee2e6;
            border-radius: 10px;
            font-size: 1em;
            width: 100px;
            transition: all 0.3s;
        }
        
        .search-box input:focus {
            outline: none;
            border-color: #667eea;
            box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
        }
        
        .btn {
            padding: 10px 25px;
            border: none;
            border-radius: 10px;
            font-size: 1em;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.3s;
            text-decoration: none;
            display: inline-block;
        }
        
        .btn-primary {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }
        
        .btn-primary:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 20px rgba(102, 126, 234, 0.3);
        }
        
        .btn-secondary {
            background: #6c757d;
            color: white;
        }
        
        .btn-secondary:hover {
            background: #5a6268;
            transform: translateY(-2px);
        }
        
        .btn:disabled {
            opacity: 0.5;
            cursor: not-allowed;
        }
        
        .viewer {
            padding: 30px;
            background: #f8f9fa;
            min-height: 600px;
            display: flex;
            flex-direction: column;
            align-items: center;
        }
        
        .page-image {
            max-width: 100%;
            height: auto;
            border: 2px solid #dee2e6;
            border-radius: 10px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.1);
            margin-bottom: 20px;
        }
        
        .page-info {
            text-align: center;
            color: #495057;
            font-size: 1.1em;
            margin-bottom: 20px;
        }
        
        .navigation {
            display: flex;
            gap: 20px;
            align-items: center;
        }
        
        .page-counter {
            font-size: 1.2em;
            font-weight: 500;
            color: #495057;
            padding: 0 20px;
        }
        
        .instructions {
            padding: 30px;
            background: #e7f3ff;
            border-left: 4px solid #007bff;
            margin: 20px;
            border-radius: 10px;
        }
        
        .instructions h2 {
            color: #007bff;
            margin-bottom: 15px;
        }
        
        .instructions ol {
            margin-left: 20px;
            line-height: 1.8;
            color: #495057;
        }
        
        .footer {
            padding: 20px;
            text-align: center;
            color: #6c757d;
            background: #f8f9fa;
            border-top: 1px solid #dee2e6;
        }
        
        @media (max-width: 768px) {
            .header h1 {
                font-size: 1.8em;
            }
            
            .info-panel {
                flex-direction: column;
                gap: 15px;
            }
            
            .controls {
                flex-direction: column;
            }
            
            .navigation {
                flex-wrap: wrap;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎨 Схема алмазной мозаики</h1>
            <p>Купон: {{.CouponCode}}</p>
        </div>
        
        <div class="info-panel">
            <div class="info-item">
                <div class="label">Всего страниц</div>
                <div class="value">{{.PageCount}}</div>
            </div>
            <div class="info-item">
                <div class="label">Количество камней</div>
                <div class="value">{{.StonesCount}}</div>
            </div>
            <div class="info-item">
                <div class="label">Текущая страница</div>
                <div class="value" id="currentPageDisplay">1</div>
            </div>
        </div>
        
        <div class="controls">
            <div class="search-box">
                <label for="pageInput">Перейти к странице:</label>
                <input 
                    type="number" 
                    id="pageInput" 
                    min="1" 
                    max="{{.PageCount}}" 
                    value="1"
                    placeholder="№"
                />
                <button class="btn btn-primary" onclick="goToPage()">Перейти</button>
            </div>
            
            <button class="btn btn-secondary" onclick="printPage()">🖨️ Печать</button>
        </div>
        
        <div class="viewer">
            <div class="page-info">
                Страница <span id="currentPage">1</span> из {{.PageCount}}
            </div>
            
            <img 
                id="pageImage" 
                class="page-image" 
                src="pages/page_001.jpg" 
                alt="Страница схемы"
                onerror="handleImageError()"
            />
            
            <div class="navigation">
                <button 
                    class="btn btn-secondary" 
                    id="prevBtn" 
                    onclick="previousPage()"
                    disabled
                >
                    ← Предыдущая
                </button>
                
                <span class="page-counter">
                    <span id="pageCounter">1</span> / {{.PageCount}}
                </span>
                
                <button 
                    class="btn btn-secondary" 
                    id="nextBtn" 
                    onclick="nextPage()"
                >
                    Следующая →
                </button>
            </div>
        </div>
        
        <div class="instructions">
            <h2>📖 Инструкция по использованию</h2>
            <ol>
                <li>Используйте кнопки навигации для перелистывания страниц</li>
                <li>Введите номер страницы в поле поиска для быстрого перехода</li>
                <li>Нажмите "Печать" для распечатки текущей страницы</li>
                <li>Все страницы схемы находятся в папке "pages"</li>
                <li>Рекомендуется печатать на бумаге формата A4</li>
            </ol>
        </div>
        
        <div class="footer">
            <p>© 2024 Алмазная мозаика. Все права защищены.</p>
            <p>Купон: {{.CouponCode}} | Создано: {{.Date}}</p>
        </div>
    </div>
    
    <script>
        let currentPage = 1;
        const totalPages = {{.PageCount}};
        
        function updatePage() {
            const pageNum = String(currentPage).padStart(3, '0');
            const imagePath = 'pages/page_' + pageNum + '.jpg';
            
            document.getElementById('pageImage').src = imagePath;
            document.getElementById('currentPage').textContent = currentPage;
            document.getElementById('currentPageDisplay').textContent = currentPage;
            document.getElementById('pageCounter').textContent = currentPage;
            document.getElementById('pageInput').value = currentPage;
            
            // Update button states
            document.getElementById('prevBtn').disabled = currentPage === 1;
            document.getElementById('nextBtn').disabled = currentPage === totalPages;
        }
        
        function nextPage() {
            if (currentPage < totalPages) {
                currentPage++;
                updatePage();
            }
        }
        
        function previousPage() {
            if (currentPage > 1) {
                currentPage--;
                updatePage();
            }
        }
        
        function goToPage() {
            const input = document.getElementById('pageInput');
            const pageNum = parseInt(input.value);
            
            if (pageNum >= 1 && pageNum <= totalPages) {
                currentPage = pageNum;
                updatePage();
            } else {
                alert('Пожалуйста, введите номер страницы от 1 до ' + totalPages);
                input.value = currentPage;
            }
        }
        
        function printPage() {
            window.print();
        }
        
        function handleImageError() {
            const img = document.getElementById('pageImage');
            img.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDAwIiBoZWlnaHQ9IjMwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KICA8cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjBmMGYwIi8+CiAgPHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSIyMCIgZmlsbD0iIzk5OTk5OSIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZG9taW5hbnQtYmFzZWxpbmU9Im1pZGRsZSI+0JjQt9C+0LHRgNCw0LbQtdC90LjQtSDQvdC1INC90LDQudC00LXQvdC+PC90ZXh0Pgo8L3N2Zz4=';
            img.alt = 'Изображение не найдено';
        }
        
        // Keyboard navigation
        document.addEventListener('keydown', function(event) {
            if (event.key === 'ArrowLeft') {
                previousPage();
            } else if (event.key === 'ArrowRight') {
                nextPage();
            }
        });
        
        // Enter key in input field
        document.getElementById('pageInput').addEventListener('keypress', function(event) {
            if (event.key === 'Enter') {
                goToPage();
            }
        });
    </script>
</body>
</html>`

	data := struct {
		CouponCode  string
		PageCount   int
		StonesCount int
		Date        string
	}{
		CouponCode:  couponCode,
		PageCount:   pageCount,
		StonesCount: stonesCount,
		Date:        "2024",
	}

	tmpl, err := template.New("index").Parse(htmlTemplate)
	if err != nil {
		return ""
	}

	var result strings.Builder
	err = tmpl.Execute(&result, data)
	if err != nil {
		return ""
	}

	return result.String()
}

// GenerateSimpleViewer generates a simpler version of the viewer
func GenerateSimpleViewer(pageCount int) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Схема мозаики</title>
    <style>
        body { font-family: Arial; text-align: center; padding: 20px; }
        img { max-width: 100%%; height: auto; margin: 20px 0; }
        button { padding: 10px 20px; margin: 0 10px; font-size: 16px; cursor: pointer; }
        input { padding: 5px; width: 60px; }
    </style>
</head>
<body>
    <h1>Схема алмазной мозаики</h1>
    <div>
        <button onclick="prev()">← Назад</button>
        <input type="number" id="page" value="1" min="1" max="%d" onchange="goTo()">
        <span> / %d</span>
        <button onclick="next()">Вперёд →</button>
    </div>
    <img id="img" src="pages/page_001.jpg">
    <script>
        var p=1,t=%d;
        function show(){
            document.getElementById('img').src='pages/page_'+String(p).padStart(3,'0')+'.jpg';
            document.getElementById('page').value=p;
        }
        function prev(){if(p>1){p--;show();}}
        function next(){if(p<t){p++;show();}}
        function goTo(){var n=parseInt(document.getElementById('page').value);if(n>=1&&n<=t){p=n;show();}}
    </script>
</body>
</html>`, pageCount, pageCount, pageCount)
}
