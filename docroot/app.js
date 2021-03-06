const radius = 20
const repellingStrength = -50
const linkLength = 100
const maxLabelLength = 25
const labelYOffset = 30

var color = d3.scaleOrdinal(d3.schemeCategory10)

var app = new Vue({
    el: '#app',

    data: {
        screen: 'loading',
        main: {
            projects: [],
            graph: { "nodes": [], "links": [] },
            svg: {},
            width: 0,
            height: 0,
            g: {},
            simulation: {},
            linkElements: [],
            nodeElements: [],
            textElements: []
        },
        showReload: false,
        error: { message: '' },
        overlay: { show: false, text: '' },
    },

    mounted: function() {
        this.initGraph()
        this.loadProjects()
    },

    methods: {
        showError: function(message) {
            this.error.message = message
            this.screen = 'error'
        },

        loadProjects: function() {
            let that = this

            d3.json("/api/projects", function(error, data) {
                if (error) {
                    that.showError(error)
                    return
                }
                if (data.error) {
                    that.showError(data.error)
                    return
                }

                that.main.projects = data
                that.screen = 'main'
            })
        },

        // copied from https://renatello.com/dynamic-drop-down-list-in-vue-js/
        selectProject: function(event) {
            let selectedIndex = event.target.options.selectedIndex - 1
            let selectedProject = this.main.projects[selectedIndex].name

            this.getGraphData(selectedProject)
        },

        reload: function() {
            let selectedIndex = this.$refs["projectSelect"].options.selectedIndex - 1
            if (selectedIndex < 0) return

            let selectedProject = this.main.projects[selectedIndex].name

            this.getGraphData(selectedProject)
        },

        initGraph: function() {
            let that = this

            this.main.svg = d3.select("svg")
                .call(d3.zoom().on("zoom", function () {
                    that.main.svg.attr("transform", d3.event.transform)
                }))

            this.main.width = +this.main.svg.node().getBoundingClientRect().width
            this.main.height = +this.main.svg.node().getBoundingClientRect().height
        },

        getGraphData: function(namespace) {
            this.reloadProject = false
            this.screen = 'loading'
            let that = this

            d3.selectAll("text").remove()
            d3.selectAll("line").remove()
            d3.selectAll("circle").remove()

            d3.json("/api/graph/" + namespace, function(error, data) {
                if (error) {
                    that.showError(error)
                    return
                }

                if (data.error) {
                    that.showError(data.error)
                    return
                }

                if (data.nodes) {
                    data.nodes.forEach(node => {
                        if (node.object && node.object.metadata && node.object.metadata.managedFields)
                            delete node.object.metadata.managedFields
                    })
                }
            
                that.main.graph = data;

                that.main.simulation = d3.forceSimulation()
                    .force("link", d3.forceLink().distance(linkLength).id(function (d) { // distance set length of links
                        return d.id
                    }))
                    .force("charge", d3.forceManyBody().strength(repellingStrength)) // strength sets repelling force
                    .force("center", d3.forceCenter(that.main.width / 2, that.main.height / 2))
                
                that.main.g = that.main.svg.append("g").attr("class", "everything")

                that.main.linkElements = that.main.svg.append("g").selectAll(".link")
                that.main.nodeElements = that.main.svg.append("g").selectAll(".nodes")
                that.main.textElements = that.main.svg.append("g").selectAll(".texts")

                that.drawGraph()

                that.main.simulation.nodes(that.main.graph.nodes).on("tick", that.ticked)
                that.main.simulation.force("link").links(that.main.graph.links)

                that.showReload = true
                that.screen = 'main'
            })
        },


        drawGraph: function() {
            let that = this

            // empty current Graph contents
            this.main.g.html('')

            this.main.linkElements = this.main.g.append("g")
                .attr("class", "links")
                .selectAll("line")
                .data(this.main.graph.links)
                .enter().append("line")
                .style("stroke-width", 3)
                .style("stroke", "grey")
        
        
            this.main.nodeElements = this.main.g.append("g")
                .attr("class", "nodes")
                .selectAll("circle")
                .data(this.main.graph.nodes)
                .enter().append("circle")
                .attr("r", radius)  // adjust this value to set radius of node
                .attr("stroke", "#fff")
                .attr('stroke-width', 21)
                .attr("id", function (d) {
                    return d.id
                })
                .attr("fill", function (d) {
                    return color(d.kind)
                })
                .on("click", this.selectNode)
                .call(d3.drag()
                    .on("start", this.dragStarted)
                    .on("drag", this.dragged)
                    .on("end", this.dragEnded))
        
            this.main.textElements = this.main.g.append("g")
                .attr("class", "texts")
                .selectAll("text")
                .data(this.main.graph.nodes)
                .enter().append("text")
                .attr("text-anchor", "end")
                .text(function (node) {
                    let label = node.kind + "/" + node.name
                    if (label.length > maxLabelLength) label = label.substring(0, maxLabelLength - 3) + "..."
                    return label
                })
                .attr("font-size", 55)
                .attr("font-family", "sans-serif")
                .attr("fill", "black")
                .attr("style", "font-weight:bold;")
                .attr("dy", labelYOffset) // vertical position of text
                .attr("text-anchor", "middle")
        },

        ticked: function() {
            let that = this

            this.main.linkElements
                .attr("x1", function (d) {
                    return d.source.x
                })
                .attr("y1", function (d) {
                    return d.source.y
                })
                .attr("x2", function (d) {
                    return d.target.x
                })
                .attr("y2", function (d) {
                    return d.target.y
                })
            this.main.nodeElements
                .attr("cx", function (d) {
                    return d.x = Math.max(radius, Math.min(that.main.width - radius, d.x))
                })
                .attr("cy", function (d) {
                    return d.y = Math.max(radius, Math.min(that.main.height - radius - labelYOffset, d.y))
                })
                .each(d => {
                    d3.select('#t_' + d.id).attr('x', d.x + 10).attr('y', d.y + 3)
                })
            this.main.textElements
                .attr('x', function (d) {
                    return d.x
                })
                .attr('y', function (d) {
                    return d.y
                })
        },

        dragStarted: function(d) {
            if (!d3.event.active) this.main.simulation.alphaTarget(0.3).restart()
            d.fx = d.x
            d.fy = d.y
        },

        dragged: function(d) {
            d.fx = d3.event.x
            d.fy = d3.event.y
        },

        dragEnded: function(d) {
            if (!d3.event.active) this.main.simulation.alphaTarget(0)
            d.fx = null
            d.fy = null
        },

        selectNode: function(d) {
            if (!d.object) {
                return
            }

            this.overlay.text = JSON.stringify(d.object, null, 2)
            this.overlay.show = true
            this.$nextTick(() => this.$refs["nodedetails"].scrollTop = 0 )
        },

        hideOverlay: function() {
            this.overlay.text = ''
            this.overlay.show = false
        },
    }
})
