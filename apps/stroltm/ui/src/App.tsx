import { useState } from 'react'
import c from "./App.module.scss"

function App() {
  const [count, setCount] = useState(0)

  return (
    <div >
      <div>
        <a href="https://vitejs.dev" target="_blank">
          <img src="/vite.svg" className={c.logo} alt="Vite logo" />
        </a>
      </div>
      <h1>Vite + React</h1>
      <div className={c.card}>
        <button onClick={() => setCount((count) => count + 1)}>
          count is {count}
        </button>
        <p>
          Edit <code>src/App.tsx</code> and save to test HMR
        </p>
      </div>
      <p className={c.read_the_docs}>
        Click on the Vite and React logos to learn more
      </p>
    </div>
  )
}

export default App
