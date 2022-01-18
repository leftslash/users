define(function (require) {
  require('lib/vue')
  const App = require('app/App')
  const app = Vue.createApp({
    components: { App },
    template:   '<App/>'
  })
  const vm = app.mount('#app')
})
