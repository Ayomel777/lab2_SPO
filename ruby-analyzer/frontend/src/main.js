import './style.css';
import { Analyze, ReadFile, SelectFile } from '../wailsjs/go/main/App';

const app = document.getElementById('app');

app.innerHTML = `
  <h1>Ruby Code Complexity Analyzer</h1>

  <div class="layout">
    <div class="panel">
      <h2>Исходный код Ruby</h2>

      <textarea
        id="source"
        placeholder="Вставь сюда Ruby-код или открой файл .rb"
      ></textarea>

      <div class="buttons">
        <button id="btn-analyze">Анализировать</button>
        <button id="btn-open">Открыть .rb</button>
      </div>
    </div>

    <div class="panel">
      <h2>Результат анализа</h2>
      <div id="result" class="result"></div>
    </div>
  </div>
`;

const $source = document.getElementById('source');
const $result = document.getElementById('result');

document.getElementById('btn-analyze').addEventListener('click', runAnalysis);

document.getElementById('btn-open').addEventListener('click', async () => {
    try {
        const path = await SelectFile();
        if (!path) return;

        const content = await ReadFile(path);
        $source.value = content;
        runAnalysis();
    } catch (e) {
        $result.innerHTML = `<p class="error">Ошибка открытия: ${escapeHtml(e)}</p>`;
    }
});

async function runAnalysis() {
    const src = $source.value;

    if (!src.trim()) {
        $result.innerHTML = '<p>Введите Ruby-код.</p>';
        return;
    }

    try {
        const res = await Analyze(src);
        render(res.gilb);
    } catch (e) {
        $result.innerHTML = `<p class="error">Ошибка: ${escapeHtml(e)}</p>`;
    }
}

function render(g) {
    const constructTable = tableHTML(
        'УПРАВЛЯЮЩИЕ КОНСТРУКЦИИ RUBY',
        g.constructs || [],
        'Конструкция',
        'Количество'
    );

    const operatorTable = tableHTML(
        'ОПЕРАТОРЫ RUBY',
        g.rubyOps || [],
        'Оператор',
        'Частота'
    );

    $result.innerHTML = `
      <h2>МЕТРИКА ДЖИЛБА</h2>

      <div class="gilb-metrics">
        <div class="gilb-card">
          <span>Абсолютная сложность CL</span>
          <b>${g.cl}</b>
        </div>

        <div class="gilb-card">
          <span>Относительная сложность cl</span>
          <b>${Number(g.relative).toFixed(2)}</b>
        </div>

        <div class="gilb-card">
          <span>Максимальная вложенность CLI</span>
          <b>${g.cli}</b>
        </div>

        <div class="gilb-card">
          <span>Количество операторов N</span>
          <b>${g.operators}</b>
        </div>
      </div>

      <div class="formula">
        <b>Расчёт относительной сложности:</b>
        <br><br>
        cl = CL / N
        <br>
        cl = ${g.cl} / ${g.operators}
        <br>
        cl ≈ ${Number(g.relative).toFixed(2)}
      </div>

      ${constructTable}
      ${operatorTable}
    `;
}

function tableHTML(title, entries, colName, colFreq) {
    if (!entries || entries.length === 0) {
        return `
          <h2>${title}</h2>
          <p class="muted">Не найдено.</p>
        `;
    }

    const rows = entries.map((e, i) => `
      <tr>
        <td>${i + 1}</td>
        <td>${escapeHtml(e.name)}</td>
        <td class="num">${e.count}</td>
      </tr>
    `).join('');

    return `
      <h2>${title}</h2>
      <table>
        <thead>
          <tr>
            <th>№</th>
            <th>${colName}</th>
            <th>${colFreq}</th>
          </tr>
        </thead>
        <tbody>${rows}</tbody>
      </table>
    `;
}

function escapeHtml(s) {
    return String(s).replace(/[&<>"']/g, c => ({
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#39;'
    })[c]);
}
