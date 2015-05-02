// Copyright 2014 The Cactus Authors. All rights reserved.

(function() {
	'use strict'

	var socket = null
	function connect() {
		socket = new WebSocket('ws://'+location.host+'/hub')

		socket.onmessage = function(event) {
			var data = JSON.parse(event.data)
			switch(data[0]) {
				case 'HELO':
					break

				case 'SYNC':
					switch(data[1]) {
						case 'contest':
							Contest.one.fetch()
							break

						case 'accounts':
							if(typeof data[2] === 'undefined') {
								Account.all.fetch()
								Account.me.fetch()
							} else {
								var acc = Account.all.get(data[2])
								if(acc) {
									acc.fetch()
								}
							}
							break

						case 'activities':
							if(Account.me.get('handle') && (Account.me.get('level') === consts.Judge || Account.me.get('level') === consts.Administrator)) {
								Activity.all.fetch({
									data: {
										cursor: Math.pow(2, 53)
									},
									remove: false
								})
							}
							break

						case 'problems':
							if(typeof data[2] === 'undefined') {
								Problem.all.fetch()
							} else {
								var prob = Problem.all.get(data[2])
								if(prob) {
									prob.fetch()
								}
							}
							break

						case 'clarifications':
							if(typeof data[2] === 'undefined') {
								Clarification.all.fetch()
							} else {
								var clar = Clarification.all.get(data[2])
								if(clar) {
									clar.fetch()
								}
							}
							break

						case 'executions':
							var exec = Execution.all.get(data[2])
							if(exec) {
								exec.fetch()
							}
							break

						case 'notifications':
							Notification.all.fetch({
								data: {
									cursor: Math.pow(2, 53)
								},
								remove: false
							})
							break

						case 'standings':
							if(typeof data[2] === 'undefined') {
								Standing.all.fetch()
							} else {
								var stnd = Standing.all.get(data[2])
								if(stnd) {
									stnd.fetch()
								}
							}
							break

						case 'standings':
							Standing.all.fetch()
							break

						case 'submissions':
							if(typeof data[2] === 'undefined') {
								Submission.all.fetch({
									data: {
										cursor: Math.pow(2, 53)
									},
									remove: false
								})
							} else {
								var subm = Submission.all.get(data[2])
								if(subm) {
									subm.fetch()
								}
							}
							break
					}
					break
			}
		}

		socket.onclose = function() {
			_.delay(function() {
				connect()
			}, 1000)
		}
	}

	var showdown = new Showdown.converter({
		extensions: [
			'github',
			'table'
		]
	})

	$(document)
	.ajaxStart(function() {
		if($.active === 1) {
			NProgress.start()
		}
	})
	.ajaxComplete(function() {
		if($.active <= 1) {
			NProgress.done()
		}
	})
	NProgress.start()

	$(window)
	.on('resize', function() {
		$('#content').css('min-height', $(window).height() - $('header').outerHeight() - $('footer').outerHeight() - 32)
	})
	.trigger('resize')
	.on('scroll', function() {

	})
	.trigger('scroll')

	$('body')
	.on('click', 'a[href^="#"]', function(event) {
		event.preventDefault()
	})
	.on('click', 'a[href^="/"][target!="_blank"]', function(event) {
		event.preventDefault()
		router.navigate($(this).attr('href'), {
			trigger: true
		})
	})
	.on('focus', 'input, select, textarea', function(event) {
		var $el = $(this)
		  , $targ = $el.next('.popover')
		if($targ.length === 0) {
			return
		}
		$el.popover({
			html: true,
			trigger: 'focus',
			content: $targ.html(),
			container: $targ.closest('.panel, .modal')
		})
		.popover('show')
	})


	function handler(func, options) {
		return function() {
			if(options.login) {
				if(!Account.me.get('handle') && location.pathname != '/login') {
					router.navigate('/login', {
						trigger: true,
						replace: true
					})
					return
				}

				if(Account.me.get('level') === consts.Participant) {
					if(!Contest.one.running() && location.pathname != '/') {
						this.navigate('/', {
							trigger: true,
							replace: true
						})
						return
					}
				}
			}

			func.apply(this, arguments)
		}
	}

	var Router = Backbone.Router.extend({
		routes: {
			'problems': handler(function() {
				content.render({
					layout: 'two'
				})

				if(Account.me.get('level') === consts.Administrator) {
					content.column('main').append($('<div class="panel-body"></div>')
						.append($('<div class="btn-toolbar"></div>')
							.append($('<div class="btn-group"></div>')
								.append($('<a class="btn btn-primary" href="/problems/create">Create</a>'))
							)
						)
					, {
						panel: true
					})
				}

				content.column('main').append(new ProblemList({
					collection: Problem.all
				}), {
					panel: true
				})
				Problem.all.fetch()
			}, {
				login: true
			}),

			'problems/create': handler(function() {
				content.render({
					layout: 'two'
				})

				content.column('main').append(new ProblemCreate(), {
					panel: true
				})
			}, {
				login: true
			}),

			'p/:slug': handler(function(slug) {
				if(slug.match(/^\d+$/)) {
					var id = parseInt(slug)
					prob = Problem.all.touch(id)
					if(!prob.get('slug')) {
						prob.once('change:slug', function() {
							router.navigate('/p/'+prob.get('slug'), {
								trigger: true,
								replace: true
							})
						})
					} else {
						router.navigate('/p/'+prob.get('slug'), {
							trigger: true,
							replace: true
						})
					}
					return
				}

				content.render({
					layout: 'two'
				})

				var prob = Problem.all.findWhere({
					slug: slug
				})
				if(!prob) {
					prob = new Problem({
						id: 'by_slug'
					})
					prob.fetch({
						data: {
							slug: slug
						}
					})
					Problem.all.add(prob)
				} else {
					prob.fetch()
				}

				content.column('main').append(new ProblemView({
					model: prob
				}), {
					panel: true
				})

				content.column('side').append(new ProblemSubmit({
					model: prob
				}), {
					panel: true
				})
			}, {
				login: true
			}),

			'p/:slug/edit': handler(function(slug) {
				content.render({
					layout: 'two'
				})

				var prob = Problem.all.findWhere({
					slug: slug
				})
				if(!prob) {
					prob = new Problem({
						id: 'by_slug'
					})
					prob.fetch({
						data: {
							slug: slug
						}
					})
					Problem.all.add(prob)
				} else {
					prob.fetch()
				}

				content.column('main').append(new ProblemEdit({
					model: prob
				}), {
					panel: true
				})
			}, {
				login: true
			}),

			'clarifications': handler(function() {
				content.render({
					layout: 'two'
				})

				var clarList = new ClarificationList({
					collection: Clarification.all
				})

				content.column('main').append($('<div class="panel-body"></div>')
					.append($('<div class="btn-toolbar"></div>')
						.append($('<div class="btn-group"></div>')
							.append($('<a class="btn btn-primary" href="#request">Request</a>')
								.on('click', function() {
									new ClarificationRequest()
									Problem.all.fetch()
								})
							)
						)
						.append($('<div class="btn-group pull-right"></div>')
							.append($('<input type="text" class="form-control">')
								.on('keyup', _.debounce(function(event) {
									clarList.filter.query = $(this).val()
									clarList.render()
								}, 125))
							)
						)
					)
				, {
					panel: true
				})

				content.column('main').append(clarList, {
					panel: true
				})
				Clarification.all.fetch()
			}, {
				login: true
			}),

			'c/:id': handler(function(id) {
				content.render({
					layout: 'two'
				})

				var clar = Clarification.all.get(id)
				if(!clar) {
					clar = new Clarification({
						id: id
					})
					Clarification.all.add(clar)
				}
				clar.fetch()

				content.column('main').append(new ClarificationView({
					model: clar
				}), {
					panel: true
				})
			}, {
				login: true
			}),

			'c/:id/edit': handler(function(id) {
				content.render({
					layout: 'two'
				})

				var clar = Clarification.all.get(id)
				if(!clar) {
					clar = new Clarification({
						id: id
					})
					Clarification.all.add(clar)
				}
				clar.fetch()

				content.column('main').append(new ClarificationEdit({
					model: clar
				}), {
					panel: true
				})

				Problem.all.fetch()
			}, {
				login: true
			}),

			'standings': handler(function() {
				content.render({
					layout: 'two'
				})

				content.column('main').append(new StandingList({
					collection: Standing.all
				}), {
					panel: true,
					table: true
				})
				Standing.all.fetch()

				Problem.all.fetch()
			}, {}),

			'statistics': handler(function() {
				content.render({
					layout: 'two'
				})

				Problem.all.fetch()
			}, {}),

			'submissions': handler(function() {
				content.render({
					layout: 'two'
				})

				var submList = new SubmissionList({
					collection: Submission.all,
					select: false
				})

				if(Account.me.get('level') === consts.Judge || Account.me.get('level') === consts.Administrator) {
					content.column('main').append(new SubmissionToolbar({
						other: submList
					}), {
						panel: true
					})
				}

				content.column('main').append(submList, {
					panel: true,
					table: true
				})
				Submission.all.fetch({
					data: {
						cursor: Math.pow(2, 53)
					},
					remove: false
				})

				content.column('side').append(new SubmissionListFilter({
					other: submList
				}), {
					panel: true
				})

				Account.all.fetch()
				Problem.all.fetch()
			}, {
				login: true
			}),

			's/:id': handler(function(id) {
				content.render({
					layout: 'two'
				})

				var subm = Submission.all.get(id)
				if(!subm) {
					subm = new Submission({
						id: id
					})
					Submission.all.add(subm)
				}
				subm.fetch()

				content.column('main').append(new SubmissionList({
					model: subm
				}), {
					panel: true,
					table: true
				})

				content.column('main').append(new SubmissionTestList({
					model: subm
				}), {
					panel: true,
					table: true
				})

				content.column('main').append(new SubmissionSource({
					model: subm
				}), {
					panel: true
				})

				if(Account.me.get('level') === consts.Judge || Account.me.get('level') === consts.Administrator) {
					content.column('side').append(new SubmissionVerdict({
						model: subm
					}), {
						panel: true
					})
				}
			}, {
				login: true
			}),

			'accounts': handler(function() {
				content.render({
					layout: 'two'
				})

				var accList = new AccountList({
					collection: Account.all
				})

				content.column('main').append($('<div class="panel-body"></div>')
					.append($('<div class="btn-toolbar"></div>')
						.append($('<div class="btn-group"></div>')
							.append($('<a class="btn btn-primary" href="/accounts/create">Create</a>'))
							.append($('<a class="btn btn-primary" href="/accounts/import">Import</a>'))
						)
						.append($('<div class="btn-group pull-right"></div>')
							.append($('<input type="text" class="form-control">')
								.on('keyup', _.debounce(function(event) {
									accList.filter.query = $(this).val()
									accList.render()
								}, 125))
							)
						)
					)
				, {
					panel: true
				})

				content.column('main').append(accList, {
					panel: true
				})
				Account.all.fetch()

				content.column('side').append(new AccountListFilter({
					other: accList
				}), {
					panel: true
				})
			}, {
				login: true
			}),

			'accounts/create': handler(function() {
				content.render({
					layout: 'two'
				})

				content.column('main').append(new AccountCreate(), {
					panel: true
				})
			}, {
				login: true
			}),

			'accounts/import': handler(function() {
				content.render({
					layout: 'two'
				})

				content.column('main').append(new AccountImport(), {
					panel: true
				})
			}, {
				login: true
			}),

			'a/:handle': handler(function(handle) {
				router.navigate('/a/'+handle+'/edit', {
					trigger: true,
					replace: true
				})
			}, {
				login: true
			}),

			'a/:handle/edit': handler(function(handle) {
				content.render({
					layout: 'two'
				})

				var acc = Account.all.findWhere({
					handle: handle
				})
				if(!acc) {
					acc = new Account({
						id: 'by_handle'
					})
					acc.fetch({
						data: {
							handle: handle
						}
					})
					Account.all.add(acc)
				} else {
					acc.fetch()
				}

				content.column('main').append(new AccountEdit({
					model: acc
				}), {
					panel: true
				})
			}, {
				login: true
			}),

			'activities': handler(function() {
				content.render({
					layout: 'two'
				})

				content.column('main').append(new ActivityList({
					collection: Activity.all
				}), {
					panel: true
				})
				Activity.all.fetch({
					data: {
						cursor: Math.pow(2, 53)
					},
					remove: false
				})
			}, {
				login: true
			}),

			'settings': handler(function() {
				content.render({
					layout: 'two'
				})

				content.column('main').append(new Settings({
					model: Contest.one
				}), {
					panel: true
				})
				Contest.one.fetch()
			}, {
				login: true
			}),

			'login': handler(function() {
				if(Account.me.get('handle')) {
					router.navigate('/', {
						trigger: true,
						replace: true
					})
					return
				}

				content.render({
					layout: 'one'
				})

				content.column('main').append(new Login())
			}, {}),

			'logout': handler(function() {
				if(!Account.me.get('handle')) {
					router.navigate('/', {
						trigger: true,
						replace: true
					})
					return
				}

				$.post('/api/logout')
				.success(function() {
					Account.me.clear()
					router.navigate('/', {
						trigger: true,
						replace: true
					})

					_.defer(function() {
						Account.all.reset()
						Activity.all.reset()
						Clarification.all.reset()
						Notification.all.reset()
						Standing.all.reset()
						Submission.all.reset()
						Execution.all.reset()
					})
				})
			}, {}),

			'': handler(function() {
				if(!Account.me.get('handle')) {
					this.navigate('/login', {
						trigger: true,
						replace: true
					})
					return
				}

				if(Account.me.get('level') === consts.Participant && (!Contest.one.started() || Contest.one.ended())) {
					content.render({
						layout: 'one'
					})

					content.column('main').append(new Splash())

					var check = _.bind(function() {
						if(Contest.one.started() && !Contest.one.ended()) {
							this.navigate('_')
							this.navigate('/', {
								trigger: true,
								replace: true
							})
							return
						}

						_.delay(function() {
							check()
						}, 1000)
					}, this)
					check()
					return
				}

				this.navigate('/problems', {
					trigger: true,
					replace: true
				})
			}, {})
		}
	})
	var router = new Router()

	router.on('route', function() {
		nav.render()
		navExtras.render()
		$(window).trigger('resize')
	})


	var Model = Backbone.Model.extend({
		initialize: function() {
			this.others = {}
		},

		resolve: function(Model, field) {
			var id = this.get(field+'Id')
			if(!id) {
				return null
			}
			var entity = Model.all.touch(id)
			  , other  = this.others[field]
			if(entity && other != entity) {
				if(other) {
					this.stopListening(other)
				}
				this.listenTo(entity, 'change', function() {
					this.trigger('change')
				}, this)
				this.others[field] = entity
			}
			return entity
		}
	})

	var Collection = Backbone.Collection.extend({
		initialize: function(options) {
			Backbone.Collection.prototype.initialize.apply(this, arguments)

			this.synced = false
			this.once('sync', _.bind(function() {
				this.synced = true
			}, this))
		},

		touch: function(id) {
			var entity = this.get(id)
			if(!entity) {
				entity = new this.model({
					id: id
				})
				this.add(entity)
				entity.fetch()
			}
			return entity
		}
	})

	var View = Backbone.View.extend({})


	var Account = Model.extend({
		defaults: {
			handle: '',
			level: '',
			name: '',
			notified: ''
		},
		urlRoot: '/api/accounts'
	})

	var Accounts = Collection.extend({
		model: Account,
		url: '/api/accounts',

		initialize: function() {
			Collection.prototype.initialize.apply(this, arguments)

			this.index = lunr(function() {
				this.ref('id')
				this.field('handle')
				this.field('name')
			})

			this.on('add', _.bind(function(clar) {
				this.listenTo(clar, 'change', function() {
					this.index.add(clar.toJSON())
				})
				this.index.add(clar.toJSON())
			}, this))
			this.on('remove', _.bind(function(clar) {
				this.index.remove(clar.toJSON())
			}, this))
		},

		comparator: function(acc) {
			var name = acc.get('name')
			  , match = name.match(/^(.*)(\d+)$/)
			if(match) {
				name = match[1]
				for(var i = match[2].length; i < 8; ++i) {
					name += '0'
				}
				name += match[2]
				return name
			}
			return name
		}
	})

	_.extend(Account, {
		all: new Accounts()
	})

	_.extend(Account, {
		me: (function() {
			var acc = new Account()
			Account.all.add(acc)
			acc.fetch({
				url: '/api/accounts/me',
				success: function() {
					main()
				}
			})
			return acc
		})()
	})


	var Activity = Model.extend({
		defaults: {
			record: ''
		},
		urlRoot:  '/api/activities'
	})

	var Activities = Collection.extend({
		model: Activity,
		url: '/api/activities',

		comparator: function(act) {
			return -act.get('id')
		}
	})

	_.extend(Activity, {
		all: new Activities()
	})


	var Contest = Model.extend({
		defaults: {
			title: '',
			header: '',
			footer: '',
			starts: '',
			length: 0
		},
		urlRoot: '/api/contests',

		started: function() {
			return moment(this.get('starts')).isBefore()
		},

		ended: function() {
			return moment(this.get('starts')).add(this.get('length'), 'minutes').isBefore()
		},

		running: function() {
			return this.started() && !this.ended()
		},

		remains: function() {
			if(Contest.one.ended()) {
				return 'Finished'
			}

			if(Contest.one.started()) {
				var remains = moment(Contest.one.get('starts')).add(Contest.one.get('length'), 'minutes').diff(moment(), 'seconds')
				return (remains/(60*60)).floor()+':'+((remains/60)%60).floor().pad(2)+':'+(remains%60).floor().pad(2)
			}

			return 'Starts '+moment(Contest.one.get('starts')).fromNow()
		}
	})
	_.extend(Contest, {
		one: (function() {
			var cnt = new Contest({
				id: 1
			})
			cnt.fetch({
				success: function() {
					main()
				}
			})
			return cnt
		})()
	})

	Contest.one.on('change', function() {
		$('title').text(this.get('title')+' | Cactus')
	})


	var Problem = Model.extend({
		defaults: {
			slug: '',
			char: '',
			title: '',
			statement: {},
			samples: [],
			notes: '',
			judge: '',
			checker: null,
			limits: {
				cpu: 0,
				memory: 0
			},
			languages: [],
			tests: [],
			scoring: ''
		},
		urlRoot:  '/api/problems'
	})

	var Problems = Collection.extend({
		model: Problem,
		url: '/api/problems'
	})

	_.extend(Problem, {
		all: new Problems()
	})


	var Clarification = Model.extend({
		defaults: {
			askerId:   0,
			problemId: 0,
			question:  '',
			response: 0,
			message: ''
		},
		urlRoot: '/api/clarifications',

		asker: function() {
			return this.resolve(Account, 'asker')
		},

		problem: function() {
			return this.resolve(Problem, 'problem')
		}
	})

	var Clarifications = Collection.extend({
		model: Clarification,
		url: '/api/clarifications',

		initialize: function() {
			Collection.prototype.initialize.apply(this, arguments)

			this.index = lunr(function() {
				this.ref('id')
				this.field('question')
				this.field('answer')
			})

			this.on('add', _.bind(function(clar) {
				this.listenTo(clar, 'change', function() {
					this.index.add(clar.toJSON())
				})
				this.index.add(clar.toJSON())
			}, this))
			this.on('remove', _.bind(function(clar) {
				this.index.remove(clar.toJSON())
			}, this))
		},

		comparator: function(subm) {
			return -subm.get('id')
		}
	})

	_.extend(Clarification, {
		all: new Clarifications()
	})


	var Notification = Model.extend({
		defaults: {
			message: ''
		},
		urlRoot:  '/api/notifications'
	})

	var Notifications = Collection.extend({
		model: Notification,
		url: '/api/notifications',

		comparator: function(act) {
			return -act.get('id')
		}
	})

	_.extend(Notification, {
		all: new Notifications()
	})


	var Standing = Model.extend({
		defaults: {
			author: {},
			score: 0,
			penalty: 0,
			attempts: {}
		}
	})

	var Standings = Collection.extend({
		model: Standing,
		url: '/api/standings'
	})

	_.extend(Standing, {
		all: new Standings()
	})


	var Submission = Model.extend({
		defaults: {
			authorId: 0,
			problemId: 0,
			language: '',
			manual: false,
			verdict: 0,
			tests: null,
			source: '',
			created: ''
		},
		urlRoot: '/api/submissions',

		author: function() {
			return this.resolve(Account, 'author')
		},

		problem: function() {
			return this.resolve(Problem, 'problem')
		}
	})

	var Submissions = Collection.extend({
		model: Submission,
		url: '/api/submissions',

		comparator: function(subm) {
			return -subm.get('id')
		}
	})

	_.extend(Submission, {
		all: new Submissions()
	})


	var Execution = Model.extend({
		defaults: {
			status: 0,
			verdict: '',
			tests: [],
			created: ''
		},
		urlRoot: '/api/executions',

		submission: function() {
			return this.resolve(Submission, 'submission')
		}
	})

	var Executions = Collection.extend({
		model: Execution,
		url: '/api/executions'
	})

	_.extend(Execution, {
		all: new Executions()
	})


	var Modal = View.extend({
		initialize: function(options) {
			_.defer(_.bind(function() {
				$('body').append($('<div class="modal fade"></div>')
					.append($('<div class="modal-dialog"></div>')
						.toggleClass('modal-lg', options.large || false)
						.append(this.$el.addClass('modal-content'))
					)
					.one('hidden.bs.modal', _.bind(function() {
						this.$el.closest('.modal').remove()
						this.remove()
					}, this))
					.modal()
					.modal('show')
				)
			}, this))
		}
	})


	var Loading = View.extend({
		id: 'loading',
		template: _.template($('#tplLoading').html()),

		initialize: function() {
			this.render()
		},

		render: function() {
			this.$el.html(this.template())
		},

		remove: function() {
			this.$('h1').animo({
				animation: 'fadeOutUp',
				duration: 0.25
			})
			this.$el.animo({
				animation: 'fadeOut',
				duration: 0.25
			}, _.bind(function() {
				View.prototype.remove.apply(this)
			}, this))
		}
	})
	var loading = new Loading()


	var Nav = View.extend({
		tagName: 'nav',
		template: _.template($('#tplNav').html()),

		initialize: function() {
			this.listenTo(Account.me, 'change:handle change:level', this.render)
			this.render()
		},

		render: function() {
			this.$el
			.addClass('nav navbar-nav')
			.html('')

			if(!Account.me.get('handle') || Account.me.get('level') === consts.Participant && (!Contest.one.started() || Contest.one.ended())) {
				return
			}

			this.$el.html(this.template())

			var frag = location.pathname
			  , fragToks = frag.split('/')
			  , fragPre = _.initial(_.first(fragToks, 3)).join('/')
			this.$('a[href="'+frag+'"], a[href="'+fragPre+'"], a[data-href-alt="'+fragPre+'"]').each(function() {
				$(this).parent()
				.addClass('active')
				.siblings().removeClass('active')
			})
		}
	})
	var nav = new Nav()

	var NavExtras = View.extend({
		tagName: 'ul',
		template: _.template($('#tplNavExtras').html()),

		initialize: function() {
			this.listenTo(Account.me, 'change:handle change:level', this.render)
			this.render()
		},

		render: function() {
			this.$el
			.addClass('dropdown-menu')
			.html(this.template({
				me: Account.me.toJSON()
			}))

			var frag = location.pathname
			  , fragToks = frag.split('/')
			  , fragPre = _.initial(_.first(fragToks, 3)).join('/')
			this.$('a[href="'+frag+'"], a[href="'+fragPre+'"], a[data-href-alt="'+fragPre+'"]').each(function() {
				$(this).parent()
				.addClass('active')
				.siblings().removeClass('active')
			})
		}
	})
	var navExtras = new NavExtras()


	var Header = View.extend({
		tagName: 'header',
		template: _.template($('#tplHeader').html()),

		events: {
			'click ul:eq(0) li:eq(0) a': function(event) {
				Account.me.save({
					notified: Notification.all.first().get('created')
				}, {
					patch: true
				})
				this.$('.navbar ul:eq(0) li:eq(0) .fa-bell').removeClass('text-danger')
			}
		},

		initialize: function() {
			this.listenTo(Contest.one, 'change', _.debounce(_.bind(this.render, this), 125))
			this.listenTo(Account.me, 'change:handle change:level change:name', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el.html(this.template({
				contest: Contest.one.toJSON(),
				me: Account.me.toJSON()
			}))

			this.$('.navbar ul:eq(0) li:eq(0)').append(new NotificationList({
				collection: Notification.all
			}).el)
			this.$('.navbar ul:eq(0) li:eq(1)').append(navExtras.el)
			this.$('.navbar ul:eq(0)').parent().prepend(nav.el)
		}
	})

	var Content = View.extend({
		tagName: 'section',
		id: 'content',

		initialize: function() {
			this.childs = []
		},

		column: function(name) {
			return {
				append: _.bind(function(view, options) {
					if(!options) {
						options = {}
					}

					var el
					if(view instanceof View) {
						el = view.el
					} else {
						el = view
					}
					this.childs.push(view)
					this.$('#'+name).append($('<div></div>')
						.toggleClass('panel panel-default', options.panel || false)
						.toggleClass('table-responsive', options.panel && options.table || false)
						.append(el)
						.animo({
							animation: options.animo ? options.animo.animation : 'fadeIn',
							duration: 0.375
						})
					)

				}, this)
			}
		},

		render: function(options) {
			_.each(this.childs, function(view) {
				if(view instanceof View) {
					if(view instanceof Timer) {
						return
					}
					view.$el.parent().remove()
					view.remove()
				} else {
					view.parent().detach()
				}
			})
			this.childs = []

			if(this.$el.is(':empty')) {
				this.$el
				.addClass('row')
				.append('<div id="main" class="col-md-12"></div>')
				.append('<div id="side" class="hidden"></div>')

				this.column('side').append(new Timer(), {
					panel: true,
					animo: {
						animation: ''
					}
				})
			}

			switch(options.layout) {
				case 'one':
					this.$('#main.col-md-9')
					.removeClass('col-md-9')
					.addClass('col-md-12')

					this.$('#side.col-md-3')
					.removeClass('col-md-3')
					.addClass('hidden')
					break

				case 'two':
					this.$('#main.col-md-12')
					.removeClass('col-md-12')
					.addClass('col-md-9')

					this.$('#side.hidden')
					.removeClass('hidden')
					.addClass('col-md-3')
					.animo({
						animation: 'fadeIn',
						duration: 0.375
					})
					break

			}
		}
	})

	var Footer = View.extend({
		tagName: 'footer',
		template: _.template($('#tplFooter').html()),

		events: {
			'click a[href="#about"]': function() {
				new About()
			}
		},

		initialize: function() {
			this.listenTo(Contest.one, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el
			.addClass('row')
			.html(this.template({
				contest: _.extend(Contest.one.toJSON(), {
					footer: $(showdown.makeHtml(Contest.one.get('footer'))).html()
				})
			}))
		}
	})


	var Login = View.extend({
		template: _.template($('#tplLogin').html()),

		events: {
			'submit form': function(event) {
				event.preventDefault()
				$.post('/api/login', {
					handle: this.$('form input[name="handle"]').val(),
					password: this.$('form input[name="password"]').val()
				})
				.success(function() {
					Account.me.fetch({
						url: '/api/accounts/me',
						success: function() {
							router.navigate('/', {
								trigger: true,
								replace: true
							})

							Notification.all.fetch({
								data: {
									cursor: Math.pow(2, 53)
								},
								remove: false
							})
						}
					})
				})
				.error(_.bind(function() {
					this.$('form').animo({
						animation: 'shake',
						duration: 0.75
					})
				}, this))
			}
		},

		initialize: function() {
			this.listenTo(Contest.one, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()

			$(window).on('resize', _.bind(function() {
				this.$el.css('marginTop', $(window).height()/2 - 192)
			}, this))
		},

		render: function(quick) {
			this.$el.html(this.template(_.extend(Contest.one.toJSON(), {
				running: Contest.one.running(),
				remains: Contest.one.remains()
			})))

			$(window).trigger('resize')
		}
	})


	var Splash = View.extend({
		template: _.template($('#tplSplash').html()),

		initialize: function() {
			this.listenTo(Contest.one, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()

			$(window).on('resize', _.bind(function() {
				this.$el.css('marginTop', $(window).height()/2 - 208)
			}, this))
		},

		render: function() {
			this.$el.html(this.template(_.extend(Contest.one.toJSON(), {
				remains: Contest.one.remains()
			})))

			$(window).trigger('resize')
		}
	})


	var Timer = View.extend({
		template: _.template($('#tplTimer').html()),

		initialize: function() {
			this.signal = _.debounce(_.bind(this.signal, this), 100)
			_.defer(this.signal)

			this.listenTo(Contest.one, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el
			.addClass('panel-body text-center')
			.html(this.template(_.extend(Contest.one.toJSON(), {
				running: Contest.one.running(),
				remains: Contest.one.remains()
			})))

			if(this.$el.parent()) {
				this.signal()
			}
		},

		signal: function() {
			this.render()
		}
	})


	var ProblemList = View.extend({
		tagName: 'ul',
		template: _.template($('#tplProblemList').html()),

		initialize: function() {
			this.listenTo(this.collection, 'add remove sync sort', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el
			.addClass('list-group')
			.html(this.template())

			if(!this.collection.synced) {
				this.$el.append($('<li class="list-group-item"></li>')
					.text('Loading..')
				)

			} else {
				this.collection.each(function(problem) {
					this.$el.append(new ProblemListItem({
						model: problem
					}).el)
				}, this)
			}
		}
	})

	var ProblemListItem = View.extend({
		tagName: 'a',
		template: _.template($('#tplProblemListItem').html()),

		initialize: function() {
			this.listenTo(this.model, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el
			.addClass('list-group-item')
			.attr('href', '/p/'+this.model.get('slug'))
			.html(this.template(this.model.toJSON()))

			this.$('.label').tooltip({
				placement: 'left'
			})
		}
	})

	var ProblemCreate = View.extend({
		template: _.template($('#tplProblemCreate').html()),

		events: {
			'submit form': function(event) {
				event.preventDefault()
				this.$('button[type=submit]').button('loading')
				Problem.all.add(this.model, {
					silent: true
				})
				this.model
				.set(_.object(_.map(this.$('[name]'), function(el) {
					var $el = $(el)
					return [$el.attr('name'), $el.val()]
				})), {
					silent: true
				})
				.save({}, {
					success: _.bind(function() {
						router.navigate('/p/'+this.model.get('slug')+'/edit', {
							trigger: true
						})
					}, this),
					error: _.bind(function() {
						this.$('button[type=submit]').button('reset')
					}, this)
				})
			}
		},

		initialize: function() {
			this.listenTo(Problem.all, 'add remove sync sort', _.debounce(_.bind(this.render, this), 125))
			this.render()

			this.model = new Problem()
		},

		render: function() {
			this.$el
			.addClass('panel-body')
			.html(this.template({
				charsUsed: Problem.all.pluck('char')
			}))
		}
	})

	var ProblemView = View.extend({
		template: _.template($('#tplProblemView').html()),

		initialize: function() {
			this.listenTo(this.model, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el
			.addClass('panel-body')
			.html(this.template(_.extend(this.model.toJSON(), {
				statement: {
					body: showdown.makeHtml(this.model.get('statement').body||''),
					input: showdown.makeHtml(this.model.get('statement').input||''),
					output: showdown.makeHtml(this.model.get('statement').output||'')
				},
				me: Account.me.toJSON()
			})))

			this.$('h2 sup').tooltip({
				placement: 'right'
			})
		}
	})

	var ProblemSubmit = View.extend({
		template: _.template($('#tplProblemSubmit').html()),

		events: {
			'submit form': function(event) {
				event.preventDefault()
				var file = this.$('input[name=source]')[0].files[0]
				  , reader = new FileReader()
				if(!file) {
					this.$('input[name=source]').animo({
						animation: 'pulse',
						duration: 0.75
					})
					return
				}
				if(file.size > this.model.get('limits').source*1024) {
					bootbox.alert('Your source size ('+numeral(file.size).format('0.0b')+') exceeds the limit by '+numeral(file.size-this.model.get('limits').source*1024).format('0.0b'))
					return
				}

				reader.onload = _.bind(function(event) {
					var subm = new Submission({
						problemId: this.model.get('id'),
						language: this.$('select[name=language]').val(),
						source: reader.result
					})
					subm.save({}, {
						success: function() {
							router.navigate('/s/'+subm.get('id'), {
								trigger: true
							})
						}
					})
				}, this)
				reader.readAsText(this.$('input[name=source]')[0].files[0])
			}
		},

		initialize: function() {
			this.listenTo(this.model, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el
			.addClass('panel-body')
			.html(this.template(this.model.toJSON()))
		}
	})

	var ProblemEdit = View.extend({
		template: _.template($('#tplProblemEdit').html()),

		events: {
			'submit form': function(event) {
				event.preventDefault()
				this.$('button[type=submit]').button('loading')
				var data = {}
				async.eachSeries(this.$('[name]'), function(el, next) {
					var $el = $(el)
					  , name = $el.attr('name')
					  , val = $el.val()
					  , step = function() {
							if(name.match(/^\w+$/)) {
								data[name] = val
							}
							if(name.match(/^\w+\[\]$/)) {
								var parts = name.match(/^(\w+)\[\]$/)
								  , key0 = parts[1]
								data[key0] = data[key0] || []
								data[key0].push(val)
							}
							if(name.match(/^\w+\[\w+\]$/)) {
								var parts = name.match(/^(\w+)\[(\w+)\]$/)
								  , key0 = parts[1]
								  , key1 = parts[2]
								  , ok = false
								data[key0] = data[key0] || {}
								data[key0][key1] = val
							}
							if(name.match(/^\w+\[\]\[\w+\]$/)) {
								var parts = name.match(/^(\w+)\[\]\[(\w+)\]$/)
								  , key0 = parts[1]
								  , key1 = parts[2]
								  , ok = false
								data[key0] = data[key0] || []
								data[key0].each(function(node) {
									if(!ok && typeof node[key1] === 'undefined') {
										node[key1] = val
										ok = true
									}
								})
								if(!ok) {
									var node = {}
									node[key1] = val
									data[key0].push(node)
								}
							}

							next()
						}

					switch($el.attr('type')) {
						case 'file':
							var file = $el[0].files[0]
							  , reader = new FileReader()
							if(!file) {
								val = null
								step()
								break
							}

							reader.onload = _.bind(function(event) {
								val = reader.result
								step()
							}, this)
							reader.readAsText(file)
							break

						case 'number':
							val = Number(val)
							step()
							break

						default:
							step()
					}
				}, _.bind(function() {
					_.each(data, function(val, key) {
						if(_.isArray(val) && typeof val[0] === 'object') {
							val.pop()
						}
					})
					this.model
					.set(data, {
						silent: true
					})
					.save({}, {
						success: _.bind(function() {
							router.navigate('/p/'+this.model.get('slug'), {
								trigger: true
							})
						}, this),
						error: _.bind(function() {
							this.$('button[type=submit]').button('reset')
						}, this)
					})
				}, this))
			},

			'click .btn-delete': function(event) {
				var $targ = $(event.target)
				bootbox.confirm($targ.data('confirm-text'), _.bind(function(okay) {
					if(!okay) {
						return
					}
					this.model.destroy({
						success: function() {
							router.navigate('/problems', {
								trigger: true
							})
						}
					})
				}, this))
			},

			'keyup input, textarea': function(event) {
				var $targ = $(event.target)
				  , $set = $targ.closest('fieldset')
				if($set.parent().hasClass('multiple')) {
					var $last = $set.parent().children().last()
					  , empty = true
					$('input, textarea', $last).each(function() {
						if($(this).val() != '') {
							empty = false
						}
					})
					if(!empty) {
						var $clone = $last.clone()
						$last.after($clone)
						$('input, textarea', $clone).val('')
					}
				}
			},

			'click .btn-reset-field-file': function(event) {
				var $elHidden = $(event.target).closest('.form-group').find('input[type=hidden]')
				  , $el = $elHidden.prev()
				$el.parent().html('').append($('<input type="file">')
					.attr('id', $el.attr('id'))
					.attr('class', $el.attr('class'))
					.attr('name', $elHidden.attr('name').replace(/Key/, ''))
				)
				$(event.target).detach()
			},

			'click .btn-delete-fieldset': function(event) {
				$(event.target).closest('fieldset').detach();
			}
		},

		initialize: function() {
			this.listenTo(this.model, 'change', _.debounce(_.bind(this.render, this), 125))
			this.listenTo(Problem.all, 'add remove sync sort', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el
			.addClass('panel-body')
			.html(this.template(_.extend(this.model.toJSON(), {
				charsUsed: Problem.all.pluck('char'),
				specs: _.flatten([this.model.get('specs'), {}]),
				samples: _.flatten([this.model.get('samples'), {}]),
				tests: _.flatten([this.model.get('tests'), {}])
			})))
		}
	})


	var ClarificationList = View.extend({
		tagName: 'ul',
		template: _.template($('#tplClarificationList').html()),

		initialize: function() {
			this.filter = {}
			this.listenTo(this.collection, 'add remove sync sort', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el
			.addClass('list-group')
			.html(this.template())

			if(!this.collection.synced) {
				this.$el.append($('<li class="list-group-item"></li>')
					.text('Loading..')
				)

			} else if(this.filter.query) {
				_.each(this.collection.index.search(this.filter.query), function(hit) {
					this.$el.append(new ClarificationListItem({
						model: this.collection.get(hit.ref)
					}).el)
				}, this)

			} else {
				this.collection.each(function(clar) {
					this.$el.append(new ClarificationListItem({
						model: clar
					}).el)
				}, this)
			}
		}
	})

	var ClarificationListItem = View.extend({
		tagName: 'a',
		template: _.template($('#tplClarificationListItem').html()),

		initialize: function() {
			this.listenTo(this.model, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el
			.addClass('list-group-item')
			.attr('href', '/c/'+this.model.get('id'))
			.html(this.template(_.extend(this.model.toJSON(), {
				asker: this.model.asker() ? this.model.asker().toJSON() : null,
				problem: this.model.problem() ? this.model.problem().toJSON() : null,
				message: $('<div></div>').html(showdown.makeHtml(this.model.get('message'))).text(),
				me: Account.me.toJSON()
			})))

			this.$('h4 small span[title]').tooltip({
				placement: 'right'
			})
		}
	})

	var ClarificationRequest = Modal.extend({
		template: _.template($('#tplClarificationRequest').html()),

		events: {
			'submit form': function(event) {
				event.preventDefault()
				this.$('button[type=submit]').button('loading')
				Clarification.all.add(this.model)
				this.model
				.set(_.object(_.map(this.$('[name]'), function(el) {
					var $el = $(el)
					  , name = $el.attr('name')
					switch(name) {
						case 'problemId':
							return [name, parseInt($el.val())]

						default:
							return [name, $el.val()]
					}
				})), {
					silent: true
				})
				.save({}, {
					success: _.bind(function() {
						this.$el.closest('.modal').modal('hide')
					}, this),
					error: _.bind(function() {
						this.$('button[type=submit]').button('reset')
					}, this)
				})
			}
		},

		initialize: function() {
			this.listenTo(Problem.all, 'add remove sync sort', _.debounce(_.bind(this.render, this), 125))
			this.render()

			this.model = new Clarification()

			Modal.prototype.initialize.call(this, {})
		},

		render: function() {
			this.$el.html(this.template({
				problems: Problem.all.toJSON()
			}))
		}
	})

	var ClarificationView = View.extend({
		template: _.template($('#tplClarificationView').html()),

		events: {
			'click .btn-answer': function(event) {
				var $el = $(event.target)
				$el.button('loading')
				this.$('[name=response]').val($el.data('value'))
			},

			'click .btn-broadcast': function(event) {
				var $el = $(event.target)
				$el.button('loading')
				this.$('[name=response]').val($el.data('value'))
			},

			'click .btn-ignore': function(event) {
				var $el = $(event.target)
				$el.button('loading')
				this.$('[name=response]').val($el.data('value'))
			},

			'submit form': function(event) {
				event.preventDefault()
				this.model
				.set(_.object(_.map(this.$('[name]'), function(el) {
					var $el = $(el)
					  , name = $el.attr('name')
					switch(name) {
						case 'response':
							return [name, parseInt($el.val())]

						default:
							return [name, $el.val()]
					}
				})), {
					silent: true
				})
				.save({}, {
					success: _.bind(function() {
						router.navigate('/clarifications', {
							trigger: true
						})
					}, this),
					error: _.bind(function() {
						this.$('button[type=submit]').button('reset')
					}, this)
				})
			}
		},

		initialize: function() {
			this.listenTo(this.model, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el
			.addClass('panel-body')
			.html(this.template(_.extend(this.model.toJSON(), {
				asker: this.model.asker() ? this.model.asker().toJSON() : null,
				problem: this.model.problem() ? this.model.problem().toJSON() : null,
				message: showdown.makeHtml(this.model.get('message')),
				me: Account.me.toJSON()
			})))

			this.$('h3 small span[title]').tooltip({
				placement: 'right'
			})
		}
	})

	var ClarificationEdit = View.extend({
		template: _.template($('#tplClarificationEdit').html()),

		events: {
			'submit form': function(event) {
				event.preventDefault()
				this.$('button[type=submit]').button('loading')
				this.model
				.set(_.object(_.map(this.$('[name]'), function(el) {
					var $el = $(el)
					  , name = $el.attr('name')
					switch(name) {
						case 'problemId':
						case 'response':
							return [name, parseInt($el.val())]

						default:
							return [name, $el.val()]
					}
				})), {
					silent: true
				})
				.save({}, {
					success: _.bind(function() {
						router.navigate('/c/'+this.model.get('id'), {
							trigger: true
						})
					}, this),
					error: _.bind(function() {
						this.$('button[type=submit]').button('reset')
					}, this)
				})
			},

			'click .btn-delete': function(event) {
				var $targ = $(event.target)
				bootbox.confirm($targ.data('confirm-text'), _.bind(function(okay) {
					if(!okay) {
						return
					}
					this.model.destroy({
						success: function() {
							router.navigate('/clarifications', {
								trigger: true
							})
						}
					})
				}, this))
			}
		},

		initialize: function() {
			this.listenTo(this.model, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()

			this.listenTo(Problem.all, 'add remove sync sort', _.debounce(_.bind(this.render, this), 125))
		},

		render: function() {
			this.$el
			.addClass('panel-body')
			.html(this.template(_.extend(this.model.toJSON(), {
				me: Account.me.toJSON(),
				problem: this.model.problem() ? this.model.problem().toJSON() : null,
				problems: Problem.all.toJSON()
			})))
		}
	})


	var StandingList = View.extend({
		tagName: 'table',
		template: _.template($('#tplStandingList').html()),

		initialize: function() {
			this.listenTo(this.collection, 'add remove sync sort', _.debounce(_.bind(this.render, this), 125))
			this.listenTo(Problem.all, 'add remove sync sort change', this.render)
			this.render()
		},

		render: function() {
			this.$el
			.addClass('table table-standing')
			.html(this.template({
				me: Account.me.toJSON(),
				problems: Problem.all.toJSON()
			}))

			this.collection.each(function(problem, i) {
				this.$('tbody').append(new StandingListItem({
					model: problem,
					rank: i+1
				}).el)
			}, this)
		}
	})

	var StandingListItem = View.extend({
		tagName: 'tr',
		template: _.template($('#tplStandingListItem').html()),

		initialize: function(options) {
			this.rank = options.rank

			this.listenTo(this.model, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el.html(this.template(_.extend(this.model.toJSON(), {
				rank: this.rank,
				me: Account.me.toJSON(),
				problems: Problem.all.toJSON()
			})))

			this.$('td .label-group').tooltip({
				placement: 'left'
			})
		}
	})


	var SubmissionToolbar = View.extend({
		template: _.template($('#tplStandingToolbar').html()),

		events: {
			'click .btn-select': function(event) {
				event.preventDefault()
				this.select = true
				this.render()
				this.other.select = true
				this.other.render()
			},

			'click .btn-cancel': function(event) {
				event.preventDefault()
				this.select = false
				this.render()
				this.other.select = false
				this.other.render()
			},

			'click .btn-reset': function(event) {
				event.preventDefault()
				async.each(this.other.$('tbody tr td:first-child input:checked').toArray().reverse(), function(el, done) {
					$.post('/api/submissions/'+parseInt($(el).val())+'/reset')
					.success(function() {
						done()
					})
				})
				this.select = false
				this.render()
				this.other.select = false
				this.other.render()
			},

			'click .btn-judge': function(event) {
				event.preventDefault()
				async.each(this.other.$('tbody tr td:first-child input:checked').toArray().reverse(), function(el, done) {
					$.post('/api/submissions/'+parseInt($(el).val())+'/judge')
					.success(function() {
						done()
					})
				})
				this.select = false
				this.render()
				this.other.select = false
				this.other.render()
			}
		},

		initialize: function(options) {
			this.other = options.other
			this.select = false
			this.render()
		},

		render: function() {
			this.$el
			.addClass('panel-body')
			.html(this.template({
				select: this.select
			}))
		}
	})

	var SubmissionList = View.extend({
		tagName: 'table',
		template: _.template($('#tplSubmissionList').html()),

		events: {
			'change thead tr th:first-child input': function(event) {
				event.preventDefault()
				var checked = $(event.target).is(':checked')
				this.$('tbody tr td:first-child input').each(function() {
					$(this).prop('checked', checked)
				})
			},

			'click tbody tr td:first-child input': function(event) {
				if(event.shiftKey && event.target != this.checks.last.target) {
					var id = $(event.target).val()
					if(this.checks.last.id < id) {
						$(event.target).closest('tr').nextUntil($(this.checks.last.target).closest('tr')).find('input').prop('checked', true)
					} else {
						$(event.target).closest('tr').prevUntil($(this.checks.last.target).closest('tr')).find('input').prop('checked', true)
					}
					return
				}

				this.checks.last = {
					target: event.target,
					id: $(event.target).val()
				}
			},

			'click a[href="#more"]': function() {
				this.collection.fetch({
					data: {
						cursor: this.collection.last().get('id')
					},
					remove: false,
					success: _.bind(function(subms, arr) {
						if(arr.length < 128) {
							this.stream = false
						}
					}, this)
				})
			}
		},

		initialize: function(options) {
			this.select = options.select
			this.checks = {
				last: {}
			}
			this.filter = {}
			this.stream = true
			if(this.model) {
				this.listenTo(this.model, 'change', _.debounce(_.bind(this.render, this), 125))
			} else {
				this.listenTo(this.collection, 'add remove sync sort', _.debounce(_.bind(this.render, this), 125))
			}
			this.render()
		},

		render: function() {
			this.$el
			.addClass('table')
			.html(this.template({
				me: Account.me.toJSON(),
				select: this.select
			}))

			if(this.model) {
				this.$('tbody').append(new SubmissionListItem({
					model: this.model,
					single: true
				}).el)
				this.$('tfoot').detach()

			} else {
				this.collection.each(function(subm) {
					this.$('tbody').append(new SubmissionListItem({
						model: subm,
						single: false,
						select: this.select,
						filter: this.filter
					}).el)
				}, this)
				if(!this.stream) {
					this.$('tfoot').detach()
				}
			}
		}
	})

	var SubmissionListItem = View.extend({
		tagName: 'tr',
		template: _.template($('#tplSubmissionListItem').html()),

		events: {
			'change td:eq(0) input': function(event) {
				this.selected = $(event.target).is(':checked')
			}
		},

		initialize: function(options) {
			this.single = options.single
			this.select = options.select
			this.filter = options.filter

			this.listenTo(this.model, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el.html(this.template(_.extend(this.model.toJSON(), {
				author: this.model.author() ? this.model.author().toJSON() : null,
				problem: this.model.problem() ? this.model.problem().toJSON() : null,
				language: _.bind(function() {
					switch(this.model.get('language')) {
						case 'c':
							return 'C'
						case 'cpp':
							return 'C++'
						case 'java':
							return 'Java'
					}
				}, this)(),
				me: Account.me.toJSON(),
				single: this.single,
				select: this.select,
				selected: this.selected
			})))

			this.$('.col-time span').tooltip({
				placement: 'right'
			})

			if(this.filter && (this.filter.author && this.model.get('authorId') != this.filter.author || this.filter.problem && this.model.get('problemId') != this.filter.problem || this.filter.language && this.model.get('language') != this.filter.language || this.filter.verdict && this.model.get('verdict') != this.filter.verdict)) {
				this.$el.hide()
			}
		}
	})

	var SubmissionListFilter = View.extend({
		template: _.template($('#tplSubmissionListFilter').html()),

		events: {
			'change select': function(event) {
				_.extend(this.other.filter, _.object(_.map(this.$('[name]'), function(el) {
					var $el = $(el)
					  , name = $el.attr('name')
					switch(name) {
						case 'author':
						case 'problem':
							return [name, parseInt($el.val())]

						default:
							return [name, $el.val()]
					}
				})))
				this.other.render()
			}
		},

		initialize: function(options) {
			this.other = options.other
			this.listenTo(Account.all, 'add remove sync sort', _.debounce(_.bind(this.render, this), 125))
			this.listenTo(Problem.all, 'add remove sync sort', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el
			.addClass('panel-body')
			.html(this.template({
				accounts: Account.all.toJSON(),
				problems: Problem.all.toJSON()
			}))
		}
	})

	var SubmissionTestList = View.extend({
		tagName: 'table',
		template: _.template($('#tplSubmissionTestList').html()),

		events: {
			'click a[href^="#compare-"]': function(event) {
				event.preventDefault()
				var $el = $(event.target)
				  , no = parseInt($el.attr('href').split('-')[1])
				new Comparator({
					leftLabel: 'Answer',
					leftUrl: this.model.problem().url()+'/tests/'+no+'/answer',
					rightLabel: 'Output',
					rightUrl: this.model.url()+'/tests/'+no+'/output'
				})
			}
		},

		initialize: function() {
			this.listenTo(this.model, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el
			.addClass('table')
			.html(this.template(_.extend(this.model.toJSON(), {
				author: this.model.author() ? this.model.author().toJSON() : null,
				problem: this.model.problem() ? this.model.problem().toJSON() : null,
				me: Account.me.toJSON()
			})))

			_.defer(_.bind(function() {
				this.$el.closest('.panel').toggle(!this.model.get('manual') && this.model.get('verdict') != 'ce')
			}, this))
		}
	})

	var SubmissionSource = View.extend({
		template: _.template($('#tplSubmissionSource').html()),

		initialize: function() {
			this.listenTo(this.model, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()

			$.get(this.model.url()+'/source')
			.success(_.bind(function(resp) {
				this.source = resp
				this.render()
			}, this))
		},

		render: function() {
			this.$el
			.addClass('panel-body')
			.html(this.template(_.extend(this.model.toJSON(), {
				source: this.model.get('language') ? hljs.highlight(this.model.get('language'), this.source||'').value : ''
			})))
		}
	})

	var SubmissionVerdict = View.extend({
		template: _.template($('#tplSubmissionVerdict').html()),

		events: {
			'click .btn-execute': function() {
				var exec = Execution.all.create({
					submissionId: this.model.get('id')
				})
				new SubmissionExecution({
					model: exec
				})
			},

			'click .btn-manual': function() {
				new SubmissionManual({
					model: this.model
				})
			}
		},

		initialize: function() {
			this.listenTo(this.model, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el
			.addClass('panel-body')
			.html(this.template(this.model.toJSON()))
		}
	})

	var SubmissionExecution = Modal.extend({
		template: _.template($('#tplSubmissionExecution').html()),

		events: {
			'submit form': function(event) {
				event.preventDefault()
				$.post('/api/executions/'+this.model.get('id')+'/apply')
				.then(_.bind(function() {
					this.$el.closest('.modal').modal('hide')
					this.model.submission().fetch()
				}, this))
			},

			'click .btn-tamper': function(event) {
				event.preventDefault()
				var view = new SubmissionManual({
					model: this.model.submission(),
					verdict: this.model.get('verdict')
				})
				_.defer(_.bind(function() {
					view.$el.closest('.modal')
					.one('hidden.bs.modal', _.bind(function() {
						if(view.applied) {
							this.$el.closest('.modal').modal('hide')
						}
					}, this))
				}, this))
			},

			'click a[href^="#compare-"]': function(event) {
				event.preventDefault()
				var $el = $(event.target)
				  , no = parseInt($el.attr('href').split('-')[1])
				new Comparator({
					leftLabel: 'Answer',
					leftUrl: this.model.submission().problem().url()+'/tests/'+no+'/answer',
					rightLabel: 'Output',
					rightUrl: this.model.url()+'/tests/'+no+'/output'
				})
			}
		},

		initialize: function() {
			this.listenTo(this.model, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()

			Modal.prototype.initialize.call(this, {
				large: true
			})
		},

		render: function() {
			this.$el.html(this.template(_.extend(this.model.toJSON(), {
				submission: this.model.submission() ? this.model.submission().toJSON() : null,
				problem: this.model.submission() && this.model.submission().problem() ? this.model.submission().problem().toJSON() : null,
			})))
		}
	})

	var SubmissionManual = Modal.extend({
		template: _.template($('#tplSubmissionManual').html()),

		events: {
			'submit form': function(event) {
				event.preventDefault()
				this.$('button[type=submit]').button('loading')
				this.model
				.set(_.object(_.map(this.$('[name]'), function(el) {
					var $el = $(el)
					  , name = $el.attr('name')
					switch(name) {
						case 'verdict':
							return [name, parseInt($el.val())]

						default:
							return [name, $el.val()]
					}
				})), {
					silent: true
				})
				.save({}, {
					success: _.bind(function() {
						this.$el.closest('.modal').modal('hide')
						this.applied = true
					}, this),
					error: _.bind(function() {
						this.$('button[type=submit]').button('reset')
					}, this)
				})
			}
		},

		initialize: function(options) {
			this.verdict = options.verdict

			this.listenTo(this.model, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()

			Modal.prototype.initialize.call(this, {})
		},

		render: function() {
			this.$el.html(this.template(_.extend(this.model.toJSON(), {
				verdict: this.verdict || this.model.get('verdict')
			})))
		}
	})


	var Comparator = Modal.extend({
		template: _.template($('#tplComparator').html()),

		events: {
			'click .btn-download-left': function(event) {
				event.preventDefault()
				open(this.leftUrl+'?download=yes', '_blank')
			},

			'click .btn-download-right': function(event) {
				event.preventDefault()
				open(this.rightUrl+'?download=yes', '_blank')
			}
		},

		initialize: function(options) {
			_.extend(this, _.pick(options, [
				'leftLabel',
				'leftUrl',
				'rightLabel',
				'rightUrl'
			]))

			this.render()

			$.get(this.leftUrl)
			.success(_.bind(function(resp) {
				this.left = resp
				this.render()
			}, this))

			$.get(this.rightUrl)
			.success(_.bind(function(resp) {
				this.right = resp
				this.render()
			}, this))

			Modal.prototype.initialize.call(this, {
				large: true
			})
		},

		render: function() {
			this.$el.html(this.template(_.extend(_.pick(this, [
				'leftLabel',
				'rightLabel',
			]), {
				left: this.left,
				right: this.right,
			})))

			this.$('pre').on('scroll', _.bind(function(event) {
				this.$('pre').not(event.target).scrollTop($(event.target).scrollTop())
			}, this))
		}
	})


	var AccountList = View.extend({
		tagName: 'ul',
		template: _.template($('#tplAccountList').html()),

		initialize: function() {
			this.filter = {}
			this.listenTo(this.collection, 'add remove sync sort', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el
			.addClass('list-group')
			.html(this.template())

			if(this.filter.query) {
				_.each(this.collection.index.search(this.filter.query), function(hit) {
					this.$el.append(new AccountListItem({
						model: this.collection.get(hit.ref),
						filter: this.filter
					}).el)
				}, this)

			} else {
				this.collection.each(function(acc) {
					this.$el.append(new AccountListItem({
						model: acc,
						filter: this.filter
					}).el)
				}, this)
			}
		}
	})

	var AccountListFilter = View.extend({
		template: _.template($('#tplAccountListFilter').html()),

		events: {
			'change select': function(event) {
				_.extend(this.other.filter, _.object(_.map(this.$('[name]'), function(el) {
					var $el = $(el)
					  , name = $el.attr('name')
					switch(name) {
						case 'level':
							return [name, parseInt($el.val())]

						default:
							return [name, $el.val()]
					}
				})))
				this.other.render()
			}
		},

		initialize: function(options) {
			this.other = options.other
			this.render()
		},

		render: function() {
			this.$el
			.addClass('panel-body')
			.html(this.template())
		}
	})

	var AccountListItem = View.extend({
		tagName: 'a',
		template: _.template($('#tplAccountListItem').html()),

		initialize: function(options) {
			this.filter = options.filter
			this.listenTo(this.model, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el
			.addClass('list-group-item')
			.attr('href', '/a/'+this.model.get('handle'))
			.html(this.template(this.model.toJSON()))

			this.$('.list-group-item-text span').tooltip({
				placement: 'right'
			})

			if(this.filter.level && this.model.get('level') != this.filter.level) {
				this.$el.hide()
			}
		}
	})

	var AccountCreate = View.extend({
		template: _.template($('#tplAccountCreate').html()),

		events: {
			'submit form': function(event) {
				event.preventDefault()
				this.$('button[type=submit]').button('loading')
				Account.all.add(this.model)
				this.model
				.set(_.object(_.map(this.$('[name]'), function(el) {
					var $el = $(el)
					  , name = $el.attr('name')
					switch(name) {
						case 'level':
							return [name, parseInt($el.val())]

						default:
							return [name, $el.val()]
					}
				})), {
					silent: true
				})
				.save({}, {
					success: function() {
						router.navigate('/accounts', {
							trigger: true
						})
					},
					error: _.bind(function() {
						this.$('button[type=submit]').button('reset')
					}, this)
				})
			}
		},

		initialize: function() {
			this.render()

			this.model = new Account()
		},

		render: function() {
			this.$el
			.addClass('panel-body')
			.html(this.template())
		}
	})

	var AccountImport = View.extend({
		template: _.template($('#tplAccountImport').html()),

		events: {
			'submit form': function(event) {
				event.preventDefault()
				var file = this.$('input[name=file]')[0].files[0]
				  , reader = new FileReader()
				if(!file) {
					this.$('input[name=file]').animo({
						animation: 'pulse',
						duration: 0.75
					})
					return
				}

				reader.onload = _.bind(function(event) {
					$.post('/api/accounts/import', reader.result)
					.then(function() {
						router.navigate('/accounts', {
							trigger: true
						})
					})
				}, this)
				reader.readAsText(file)
			}
		},

		initialize: function() {
			this.render()
		},

		render: function() {
			this.$el
			.addClass('panel-body')
			.html(this.template())
		}
	})

	var AccountEdit = View.extend({
		template: _.template($('#tplAccountEdit').html()),

		events: {
			'submit form': function(event) {
				event.preventDefault()
				this.$('button[type=submit]').button('loading')
				this.model
				.set(_.object(_.map(this.$('[name]'), function(el) {
					var $el = $(el)
					  , name = $el.attr('name')
					switch(name) {
						case 'level':
							return [name, parseInt($el.val())]

						default:
							return [name, $el.val()]
					}
				})), {
					silent: true
				})
				.save({}, {
					success: _.bind(function() {
						router.navigate('/accounts', {
							trigger: true
						})
					}, this),
					error: _.bind(function() {
						this.$('button[type=submit]').button('reset')
					}, this)
				})
			},

			'click .btn-delete': function(event) {
				var $targ = $(event.target)
				bootbox.confirm($targ.data('confirm-text'), _.bind(function(okay) {
					if(!okay) {
						return
					}
					this.model.destroy({
						success: function() {
							router.navigate('/accounts', {
								trigger: true
							})
						}
					})
				}, this))
			}
		},

		initialize: function() {
			this.listenTo(this.model, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el
			.addClass('panel-body')
			.html(this.template(_.extend(this.model.toJSON(), {
				me: Account.me.toJSON()
			})))
		}
	})


	var ActivityList = View.extend({
		tagName: 'ul',
		template: _.template($('#tplActivityList').html()),

		events: {
			'click a[href="#more"]': function() {
				this.collection.fetch({
					data: {
						cursor: this.collection.last().get('id')
					},
					remove: false,
					success: _.bind(function(acts, arr) {
						if(arr.length < 256) {
							this.stream = false
						}
					}, this)
				})
			}
		},

		initialize: function() {
			this.stream = true
			this.listenTo(this.collection, 'add remove sync sort', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el
			.addClass('list-group')
			.html(this.template())

			this.collection.each(function(act) {
				this.$el.append(new ActivityListItem({
					model: act
				}).el)
			}, this)
			this.$el.append($('<a class="list-group-item small" href="#more"></a>')
				.text('More..')
			)

			if(!this.stream) {
				this.$('a[href="#more"]').detach()
			}
		}
	})

	var ActivityListItem = View.extend({
		tagName: 'li',
		template: _.template($('#tplActivityListItem').html()),

		initialize: function() {
			this.listenTo(this.model, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el
			.addClass('list-group-item')
			.html(this.template(_.extend(this.model.toJSON(), {
				record: _.bind(function() {
					return this.model.get('record')
					.replace(/\b(account (\d+))\b/g, '<a href="/a/$2">$1</a>')
					.replace(/\b(clarification (\d+))\b/g, '<a href="/c/$2">$1</a>')
					.replace(/\b(execution (\d+))\b/g, '<a href="#execution-$2">$1</a>')
					.replace(/\b(problem (\d+))\b/g, '<a href="/p/$2">$1</a>')
					.replace(/\b(submission (\d+))\b/g, '<a href="/s/$2">$1</a>')
				}, this)()
			})))

			this.$('.list-group-item-text span').tooltip({
				placement: 'left'
			})
		}
	})


	var NotificationList = View.extend({
		tagName: 'ul',
		template: _.template($('#tplNotificationList').html()),

		events: {
			'click a[href="#more"]': function(event) {
				event.preventDefault()
				event.stopPropagation()
				this.collection.fetch({
					data: {
						cursor: this.collection.last().get('id')
					},
					remove: false,
					success: _.bind(function(notifs, arr) {
						if(arr.length < 32) {
							this.stream = false
						}
					}, this)
				})
			}
		},

		initialize: function() {
			this.stream = true
			this.listenTo(this.collection, 'add remove sync sort', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el
			.addClass('dropdown-menu list-group')
			.html(this.template())

			this.collection.each(function(act) {
				this.$el.append(new NotificationListItem({
					model: act
				}).el)
			}, this)
			this.$el.append($('<a class="list-group-item small" href="#more"></a>')
				.text('More..')
			)

			_.defer(_.bind(function() {
				if(this.collection.first() && moment(this.collection.first().get('created')).isAfter(moment(Account.me.get('notified')))) {
					this.$el.closest('.dropdown').find('.fa-bell')
					.addClass('text-danger')
					.animo({
						animation: 'tada',
						iterate: 3
					})

				} else {
					this.$el.closest('.dropdown').find('.fa-bell').removeClass('text-danger')
				}
			}, this))

			if(!this.stream) {
				this.$('a[href="#more"]').detach()
			}
		}
	})

	var NotificationListItem = View.extend({
		tagName: 'span',
		template: _.template($('#tplNotificationListItem').html()),

		initialize: function() {
			this.listenTo(this.model, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()
		},

		render: function() {
			this.$el
			.addClass('list-group-item')
			.html(this.template(_.extend(this.model.toJSON(), {
				message: _.bind(function() {
					return this.model.get('message')
					.replace(/\b(account (\d+))\b/gi, '<a href="/a/$2">$1</a>')
					.replace(/\b(clarification (\d+))\b/gi, '<a href="/c/$2">$1</a>')
					.replace(/\b(execution (\d+))\b/gi, '<a href="#execution-$2">$1</a>')
					.replace(/\b(problem (\d+))\b/gi, '<a href="/p/$2">$1</a>')
					.replace(/\b(submission (\d+))\b/gi, '<a href="/s/$2">$1</a>')
				}, this)()
			})))
		}
	})


	var Settings = View.extend({
		template: _.template($('#tplSettings').html()),

		events: {
			'submit form': function(event) {
				event.preventDefault()
				this.$('button[type=submit]').button('loading')
				this.model
				.set(_.object(_.map(this.$('[name]'), function(el) {
					var $el = $(el)
					  , name = $el.attr('name')
					switch(name) {
						case 'length':
							return [name, parseInt($el.val())]

						default:
							return [name, $el.val()]
					}
				})), {
					silent: true
				})
				.save({}, {
					success: _.bind(function() {
						this.$('button[type=submit]').button('reset')
					}, this),
					error: _.bind(function() {
						this.$('button[type=submit]').button('reset')
					}, this)
				})
			},

			'change input[name="starts"]': function(event) {
				var $targ = $(event.target)
				$targ.val(Date.create($targ.val()).format(Date.ISO8601_DATETIME))
			}
		},

		initialize: function() {
			this.render()
		},

		render: function() {
			this.$el
			.addClass('panel-body')
			.html(this.template(this.model.toJSON()))
		}
	})


	var About = Modal.extend({
		template: _.template($('#tplAbout').html()),

		events: {
			'click a[href="#credits"]': function() {
				$.getJSON('/assets/json/credits.json')
				.success(function(credits) {
					new Credits({
						collection: new Backbone.Collection(credits)
					})
				})
			},

			'click a[href="#terms"]': function() {
				new Terms()
			}
		},

		initialize: function() {
			this.render()

			Modal.prototype.initialize.call(this, {})
		},

		render: function() {
			this.$el.html(this.template())
		}
	})

	var Credits = Modal.extend({
		template: _.template($('#tplCredits').html()),

		initialize: function() {
			this.render()

			Modal.prototype.initialize.call(this, {
				large: true
			})
		},

		render: function() {
			this.$el
			.html(this.template())
			.find('.panel-group')
			.append(this.collection.map(function(credit) {
				return $('<div></div>').append($('<div class="panel panel-default"></div>')
					.append($('<div class="panel-heading"></div>')
						.append($('<h4 class="panel-title"></h4>')
							.append($('<a data-toggle="collapse" data-parent="#credits"></a>')
								.attr('href', '#credits-'+credit.get('name')
									.toLowerCase()
									.replace(/[^a-z0-9]+/g, '')
								)
								.text(credit.get('name'))
							)
						)
					)
					.append($('<div class="panel-collapse collapse"></div>')
						.attr('id', 'credits-'+credit.get('name')
							.toLowerCase()
							.replace(/[^a-z0-9]+/g, '')
						)
						.append($('<div class="panel-body"></div>')
							.append($('<p></p>')
								.text(credit.get('description'))
								.append($('<small> &mdash; </small>')
									.append($('<a target="_blank"></a>')
										.attr('href', credit.get('homepage'))
										.text(credit.get('homepage')
											.replace(/^http:\/\//, '')
											.replace(/\/$/, '')
										)
									)
								)
							)
							.append($('<p></p>')
								.append($('<a href="#credits-'+credit.get('name')+'-license">License</a>')
									.on('click', function() {
										new License({
											model: credit
										})
									})
								)
							)
						)
					)
				).children()
			}))
		}
	})

	var Terms = Modal.extend({
		template: _.template($('#tplTerms').html()),

		initialize: function() {
			this.render()

			Modal.prototype.initialize.call(this, {
				large: true
			})
		},

		render: function() {
			this.$el.html(this.template())
		}
	})

	var License = Modal.extend({
		template: _.template($('#tplLicense').html()),

		initialize: function() {
			this.listenTo(this.model, 'change', _.debounce(_.bind(this.render, this), 125))
			this.render()

			Modal.prototype.initialize.call(this, {
				large: true
			})
		},

		render: function() {
			this.$el.html(this.template(this.model.toJSON()))
		}
	})


	var header = new Header()
	  , content = new Content()
	  , footer = new Footer()

	$('body')
	.append(loading.el)
	.append(header.el)
	.append(content.el)
	.append(footer.el)


	var main = _.after(2, function() {
		$.ajaxSetup({
			cache: false
		})

		Backbone.history.start({
			pushState: true
		})
		_.delay(function() {
			loading.remove()
		}, 750)

		if(Account.me.get('handle')) {
			Notification.all.fetch({
				data: {
					cursor: Math.pow(2, 53)
				},
				remove: false
			})
		}

		connect()
	})

	var tick = function() {
		$('span[data-time]').each(function() {
			$(this).text(moment($(this).data('time')).fromNow())
		})
		_.delay(tick, 10000)
	}
	tick()
})()
