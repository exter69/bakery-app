<template>
  <div class="bg-gray-50">
    <!-- Mobile menu -->
    <TransitionRoot as="template" :show="open">
      <Dialog class="relative z-40 lg:hidden" @close="open = false">
        <TransitionChild as="template" enter="transition-opacity ease-linear duration-300" enter-from="opacity-0" enter-to="" leave="transition-opacity ease-linear duration-300" leave-from="" leave-to="opacity-0">
          <div class="fixed inset-0 bg-black/25"></div>
        </TransitionChild>
        <div class="fixed inset-0 z-40 flex">
          <TransitionChild as="template" enter="transition ease-in-out duration-300 transform" enter-from="-translate-x-full" enter-to="translate-x-0" leave="transition ease-in-out duration-300 transform" leave-from="translate-x-0" leave-to="-translate-x-full">
            <DialogPanel class="relative flex w-full max-w-xs flex-col overflow-y-auto bg-white pb-12 shadow-xl">
              <div class="flex px-4 pt-5 pb-2">
                <button type="button" class="relative -m-2 inline-flex items-center justify-center rounded-md p-2 text-gray-400" @click="open = false">
                  <span class="absolute -inset-0.5"></span>
                  <span class="sr-only">Close menu</span>
                  <XMarkIcon class="size-6" aria-hidden="true" />
                </button>
              </div>

              <!-- Links -->
              <TabGroup as="div" class="mt-2">
                <div class="border-b border-gray-200">
                  <TabList class="-mb-px flex space-x-8 px-4">
                    <Tab as="template" v-for="category in navigation.categories" :key="category.name" v-slot="{ selected }">
                      <button :class="[selected ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-gray-900', 'flex-1 border-b-2 px-1 py-4 text-base font-medium whitespace-nowrap']">{{ category.name }}</button>
                    </Tab>
                  </TabList>
                </div>
                <TabPanels as="template">
                  <TabPanel v-for="category in navigation.categories" :key="category.name" class="space-y-10 px-4 pt-10 pb-8">
                    <div class="grid grid-cols-2 gap-x-4">
                      <div v-for="item in category.featured" :key="item.name" class="group relative text-sm">
                        <img :src="item.imageSrc" :alt="item.imageAlt" class="aspect-square w-full rounded-lg bg-gray-100 object-cover group-hover:opacity-75" />
                        <a :href="item.href" class="mt-6 block font-medium text-gray-900">
                          <span class="absolute inset-0 z-10" aria-hidden="true"></span>
                          {{ item.name }}
                        </a>
                        <p aria-hidden="true" class="mt-1">Shop now</p>
                      </div>
                    </div>
                    <div v-for="section in category.sections" :key="section.name">
                      <p :id="`${category.id}-${section.id}-heading-mobile`" class="font-medium text-gray-900">{{ section.name }}</p>
                      <ul role="list" :aria-labelledby="`${category.id}-${section.id}-heading-mobile`" class="mt-6 flex flex-col space-y-6">
                        <li v-for="item in section.items" :key="item.name" class="flow-root">
                          <a :href="item.href" class="-m-2 block p-2 text-gray-500">{{ item.name }}</a>
                        </li>
                      </ul>
                    </div>
                  </TabPanel>
                </TabPanels>
              </TabGroup>

              <div class="space-y-6 border-t border-gray-200 px-4 py-6">
                <div v-for="page in navigation.pages" :key="page.name" class="flow-root">
                  <a :href="page.href" class="-m-2 block p-2 font-medium text-gray-900">{{ page.name }}</a>
                </div>
              </div>

              <div class="space-y-6 border-t border-gray-200 px-4 py-6">
                <div class="flow-root">
                  <a href="#" class="-m-2 block p-2 font-medium text-gray-900">Sign in</a>
                </div>
                <div class="flow-root">
                  <a href="#" class="-m-2 block p-2 font-medium text-gray-900">Create account</a>
                </div>
              </div>

              <div class="border-t border-gray-200 px-4 py-6">
                <a href="#" class="-m-2 flex items-center p-2">
                  <img src="https://tailwindcss.com/plus-assets/img/flags/flag-canada.svg" alt="" class="block h-auto w-5 shrink-0" />
                  <span class="ml-3 block text-base font-medium text-gray-900">CAD</span>
                  <span class="sr-only">, change currency</span>
                </a>
              </div>
            </DialogPanel>
          </TransitionChild>
        </div>
      </Dialog>
    </TransitionRoot>

    <header class="relative bg-white">
      <p class="flex h-10 items-center justify-center bg-indigo-600 px-4 text-sm font-medium text-white sm:px-6 lg:px-8">Get free delivery on orders over $100</p>

      <nav aria-label="Top" class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div class="border-b border-gray-200">
          <div class="flex h-16 items-center">
            <button type="button" class="relative rounded-md bg-white p-2 text-gray-400 lg:hidden" @click="open = true">
              <span class="absolute -inset-0.5"></span>
              <span class="sr-only">Open menu</span>
              <Bars3Icon class="size-6" aria-hidden="true" />
            </button>

            <!-- Logo -->
            <div class="ml-4 flex lg:ml-0">
              <a href="#">
                <span class="sr-only">Your Company</span>
                <img class="h-8 w-auto" src="https://tailwindcss.com/plus-assets/img/logos/mark.svg?color=indigo&shade=600" alt="" />
              </a>
            </div>

            <!-- Flyout menus -->
            <PopoverGroup class="hidden lg:ml-8 lg:block lg:self-stretch">
              <div class="flex h-full space-x-8">
                <Popover v-for="category in navigation.categories" :key="category.name" class="flex" v-slot="{ open }">
                  <div class="relative flex">
                    <PopoverButton :class="[open ? 'text-indigo-600' : 'text-gray-700 hover:text-gray-800', 'relative flex items-center justify-center text-sm font-medium transition-colors duration-200 ease-out']">
                      {{ category.name }}
                      <span :class="[open ? 'bg-indigo-600' : '', 'absolute inset-x-0 -bottom-px z-30 h-0.5 transition duration-200 ease-out']" aria-hidden="true"></span>
                    </PopoverButton>
                  </div>
                  <transition enter-active-class="transition ease-out duration-200" enter-from-class="opacity-0" enter-to-class="" leave-active-class="transition ease-in duration-150" leave-from-class="" leave-to-class="opacity-0">
                    <PopoverPanel class="absolute inset-x-0 top-full z-20 w-full bg-white text-sm text-gray-500">
                      <!-- Presentational element used to render the bottom shadow, if we put the shadow on the actual panel it pokes out the top, so we use this shorter element to hide the top of the shadow -->
                      <div class="absolute inset-0 top-1/2 bg-white shadow-sm" aria-hidden="true"></div>
                      <div class="relative bg-white">
                        <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
                          <div class="grid grid-cols-2 gap-x-8 gap-y-10 py-16">
                            <div class="col-start-2 grid grid-cols-2 gap-x-8">
                              <div v-for="item in category.featured" :key="item.name" class="group relative text-base sm:text-sm">
                                <img :src="item.imageSrc" :alt="item.imageAlt" class="aspect-square w-full rounded-lg bg-gray-100 object-cover group-hover:opacity-75" />
                                <a :href="item.href" class="mt-6 block font-medium text-gray-900">
                                  <span class="absolute inset-0 z-10" aria-hidden="true"></span>
                                  {{ item.name }}
                                </a>
                                <p aria-hidden="true" class="mt-1">Shop now</p>
                              </div>
                            </div>
                            <div class="row-start-1 grid grid-cols-3 gap-x-8 gap-y-10 text-sm">
                              <div v-for="section in category.sections" :key="section.name">
                                <p :id="`${section.name}-heading`" class="font-medium text-gray-900">{{ section.name }}</p>
                                <ul role="list" :aria-labelledby="`${section.name}-heading`" class="mt-6 space-y-6 sm:mt-4 sm:space-y-4">
                                  <li v-for="item in section.items" :key="item.name" class="flex">
                                    <a :href="item.href" class="hover:text-gray-800">{{ item.name }}</a>
                                  </li>
                                </ul>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </PopoverPanel>
                  </transition>
                </Popover>
                <a v-for="page in navigation.pages" :key="page.name" :href="page.href" class="flex items-center text-sm font-medium text-gray-700 hover:text-gray-800">{{ page.name }}</a>
              </div>
            </PopoverGroup>

            <div class="ml-auto flex items-center">
              <div class="hidden lg:flex lg:flex-1 lg:items-center lg:justify-end lg:space-x-6">
                <a href="#" class="text-sm font-medium text-gray-700 hover:text-gray-800">Sign in</a>
                <span class="h-6 w-px bg-gray-200" aria-hidden="true"></span>
                <a href="#" class="text-sm font-medium text-gray-700 hover:text-gray-800">Create account</a>
              </div>

              <div class="hidden lg:ml-8 lg:flex">
                <a href="#" class="flex items-center text-gray-700 hover:text-gray-800">
                  <img src="https://tailwindcss.com/plus-assets/img/flags/flag-canada.svg" alt="" class="block h-auto w-5 shrink-0" />
                  <span class="ml-3 block text-sm font-medium">CAD</span>
                  <span class="sr-only">, change currency</span>
                </a>
              </div>

              <!-- Search -->
              <div class="flex lg:ml-6">
                <a href="#" class="p-2 text-gray-400 hover:text-gray-500">
                  <span class="sr-only">Search</span>
                  <MagnifyingGlassIcon class="size-6" aria-hidden="true" />
                </a>
              </div>

              <!-- Cart -->
              <div class="ml-4 flow-root lg:ml-6">
                <a href="#" class="group -m-2 flex items-center p-2">
                  <ShoppingBagIcon class="size-6 shrink-0 text-gray-400 group-hover:text-gray-500" aria-hidden="true" />
                  <span class="ml-2 text-sm font-medium text-gray-700 group-hover:text-gray-800">0</span>
                  <span class="sr-only">items in cart, view bag</span>
                </a>
              </div>
            </div>
          </div>
        </div>
      </nav>
    </header>

    <main>
      <!-- Product -->
      <div class="bg-white">
        <div class="mx-auto max-w-2xl px-4 pt-16 pb-24 sm:px-6 sm:pt-24 sm:pb-32 lg:grid lg:max-w-7xl lg:grid-cols-2 lg:gap-x-8 lg:px-8">
          <!-- Product details -->
          <div class="lg:max-w-lg lg:self-end">
            <nav aria-label="Breadcrumb">
              <ol role="list" class="flex items-center space-x-2">
                <li v-for="(breadcrumb, breadcrumbIdx) in product.breadcrumbs" :key="breadcrumb.id">
                  <div class="flex items-center text-sm">
                    <a :href="breadcrumb.href" class="font-medium text-gray-500 hover:text-gray-900">{{ breadcrumb.name }}</a>
                    <svg v-if="breadcrumbIdx !== product.breadcrumbs.length - 1" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true" class="ml-2 size-5 shrink-0 text-gray-300">
                      <path d="M5.555 17.776l8-16 .894.448-8 16-.894-.448z" />
                    </svg>
                  </div>
                </li>
              </ol>
            </nav>

            <div class="mt-4">
              <h1 class="text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl">{{ product.name }}</h1>
            </div>

            <section aria-labelledby="information-heading" class="mt-4">
              <h2 id="information-heading" class="sr-only">Product information</h2>

              <div class="flex items-center">
                <p class="text-lg text-gray-900 sm:text-xl">{{ product.price }}</p>

                <div class="ml-4 border-l border-gray-300 pl-4">
                  <h2 class="sr-only">Reviews</h2>
                  <div class="flex items-center">
                    <div>
                      <div class="flex items-center">
                        <StarIcon v-for="rating in [0, 1, 2, 3, 4]" :key="rating" :class="[reviews.average > rating ? 'text-yellow-400' : 'text-gray-300', 'size-5 shrink-0']" aria-hidden="true" />
                      </div>
                      <p class="sr-only">{{ reviews.average }} out of 5 stars</p>
                    </div>
                    <p class="ml-2 text-sm text-gray-500">{{ reviews.totalCount }} reviews</p>
                  </div>
                </div>
              </div>

              <div class="mt-4 space-y-6">
                <p class="text-base text-gray-500">{{ product.description }}</p>
              </div>

              <div class="mt-6 flex items-center">
                <CheckIcon class="size-5 shrink-0 text-green-500" aria-hidden="true" />
                <p class="ml-2 text-sm text-gray-500">In stock and ready to ship</p>
              </div>
            </section>
          </div>

          <!-- Product image -->
          <div class="mt-10 lg:col-start-2 lg:row-span-2 lg:mt-0 lg:self-center">
            <img :src="product.imageSrc" :alt="product.imageAlt" class="aspect-square w-full rounded-lg object-cover" />
          </div>

          <!-- Product form -->
          <div class="mt-10 lg:col-start-1 lg:row-start-2 lg:max-w-lg lg:self-start">
            <section aria-labelledby="options-heading">
              <h2 id="options-heading" class="sr-only">Product options</h2>

              <form>
                <div class="sm:flex sm:justify-between">
                  <!-- Size selector -->
                  <fieldset>
                    <legend class="block text-sm font-medium text-gray-700">Size</legend>
                    <div class="mt-1 grid grid-cols-1 gap-4 sm:grid-cols-2">
                      <label v-for="size in product.sizes" :key="size.name" :aria-label="size.name" :aria-description="size.description" class="group relative flex rounded-lg border border-gray-300 bg-white p-4 has-checked:outline-2 has-checked:-outline-offset-2 has-checked:outline-indigo-600 has-focus-visible:outline-3 has-focus-visible:-outline-offset-1 has-disabled:border-gray-400 has-disabled:bg-gray-200 has-disabled:opacity-25">
                        <input type="radio" name="size" :value="size.name" :checked="size === product.sizes[0]" class="absolute inset-0 appearance-none focus:outline-none" />
                        <div class="flex-1">
                          <span class="block text-base font-medium text-gray-900">{{ size.name }}</span>
                          <span class="mt-1 block text-sm text-gray-500">{{ size.description }}</span>
                        </div>
                        <CheckCircleIcon class="invisible size-5 text-indigo-600 group-has-checked:visible" aria-hidden="true" />
                      </label>
                    </div>
                  </fieldset>
                </div>
                <div class="mt-4">
                  <a href="#" class="group inline-flex text-sm text-gray-500 hover:text-gray-700">
                    <span>What size should I buy?</span>
                    <QuestionMarkCircleIcon class="ml-2 size-5 shrink-0 text-gray-400 group-hover:text-gray-500" aria-hidden="true" />
                  </a>
                </div>
                <div class="mt-10">
                  <button type="submit" class="flex w-full items-center justify-center rounded-md border border-transparent bg-indigo-600 px-8 py-3 text-base font-medium text-white hover:bg-indigo-700 focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:ring-offset-gray-50 focus:outline-hidden">Add to bag</button>
                </div>
                <div class="mt-6 text-center">
                  <a href="#" class="group inline-flex text-base font-medium">
                    <ShieldCheckIcon class="mr-2 size-6 shrink-0 text-gray-400 group-hover:text-gray-500" aria-hidden="true" />
                    <span class="text-gray-500 hover:text-gray-700">Lifetime Guarantee</span>
                  </a>
                </div>
              </form>
            </section>
          </div>
        </div>
      </div>

      <div class="mx-auto max-w-2xl px-4 py-24 sm:px-6 sm:py-32 lg:max-w-7xl lg:px-8">
        <!-- Details section -->
        <section aria-labelledby="details-heading">
          <div class="flex flex-col items-center text-center">
            <h2 id="details-heading" class="text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl">The Fine Details</h2>
            <p class="mt-3 max-w-3xl text-lg text-gray-600">Our patented padded snack sleeve construction protects your favorite treats from getting smooshed during all-day adventures, long shifts at work, and tough travel schedules.</p>
          </div>

          <div class="mt-16 grid grid-cols-1 gap-y-16 lg:grid-cols-2 lg:gap-x-8">
            <div>
              <img src="https://tailwindcss.com/plus-assets/img/ecommerce-images/product-page-04-detail-product-shot-01.jpg" alt="Drawstring top with elastic loop closure and textured interior padding." class="aspect-3/2 w-full rounded-lg object-cover" />
              <p class="mt-8 text-base text-gray-500">The 20L model has enough space for 370 candy bars, 6 cylinders of chips, 1,220 standard gumballs, or any combination of on-the-go treats that your heart desires. Yes, we did the math.</p>
            </div>
            <div>
              <img src="https://tailwindcss.com/plus-assets/img/ecommerce-images/product-page-04-detail-product-shot-02.jpg" alt="Front zipper pouch with included key ring." class="aspect-3/2 w-full rounded-lg object-cover" />
              <p class="mt-8 text-base text-gray-500">Up your snack organization game with multiple compartment options. The quick-access stash pouch is ready for even the most unexpected snack attacks and sharing needs.</p>
            </div>
          </div>
        </section>

        <!-- Policies section -->
        <section aria-labelledby="policy-heading" class="mt-16 lg:mt-24">
          <h2 id="policy-heading" class="sr-only">Our policies</h2>
          <div class="grid grid-cols-1 gap-y-12 sm:grid-cols-2 sm:gap-x-6 lg:grid-cols-4 lg:gap-x-8">
            <div v-for="policy in policies" :key="policy.name">
              <img :src="policy.imageSrc" alt="" class="h-24 w-auto" />
              <h3 class="mt-6 text-base font-medium text-gray-900">{{ policy.name }}</h3>
              <p class="mt-3 text-base text-gray-500">{{ policy.description }}</p>
            </div>
          </div>
        </section>
      </div>

      <section aria-labelledby="reviews-heading" class="bg-white">
        <div class="mx-auto max-w-2xl px-4 py-24 sm:px-6 lg:grid lg:max-w-7xl lg:grid-cols-12 lg:gap-x-8 lg:px-8 lg:py-32">
          <div class="lg:col-span-4">
            <h2 id="reviews-heading" class="text-2xl font-bold tracking-tight text-gray-900">Customer Reviews</h2>

            <div class="mt-3 flex items-center">
              <div>
                <div class="flex items-center">
                  <StarIcon v-for="rating in [0, 1, 2, 3, 4]" :key="rating" :class="[reviews.average > rating ? 'text-yellow-400' : 'text-gray-300', 'size-5 shrink-0']" aria-hidden="true" />
                </div>
                <p class="sr-only">{{ reviews.average }} out of 5 stars</p>
              </div>
              <p class="ml-2 text-sm text-gray-900">Based on {{ reviews.totalCount }} reviews</p>
            </div>

            <div class="mt-6">
              <h3 class="sr-only">Review data</h3>

              <dl class="space-y-3">
                <div v-for="count in reviews.counts" :key="count.rating" class="flex items-center text-sm">
                  <dt class="flex flex-1 items-center">
                    <p class="w-3 font-medium text-gray-900">{{ count.rating }}<span class="sr-only"> star reviews</span></p>
                    <div aria-hidden="true" class="ml-1 flex flex-1 items-center">
                      <StarIcon :class="[count.count > 0 ? 'text-yellow-400' : 'text-gray-300', 'size-5 shrink-0']" aria-hidden="true" />

                      <div class="relative ml-3 flex-1">
                        <div class="h-3 rounded-full border border-gray-200 bg-gray-100"></div>
                        <div v-if="count.count > 0" class="absolute inset-y-0 rounded-full border border-yellow-400 bg-yellow-400" :style="{ width: `calc(${count.count} / ${reviews.totalCount} * 100%)` }"></div>
                      </div>
                    </div>
                  </dt>
                  <dd class="ml-3 w-10 text-right text-sm text-gray-900 tabular-nums">{{ Math.round((count.count / reviews.totalCount) * 100) }}%</dd>
                </div>
              </dl>
            </div>

            <div class="mt-10">
              <h3 class="text-lg font-medium text-gray-900">Share your thoughts</h3>
              <p class="mt-1 text-sm text-gray-600">If you’ve used this product, share your thoughts with other customers</p>

              <a href="#" class="mt-6 inline-flex w-full items-center justify-center rounded-md border border-gray-300 bg-white px-8 py-2 text-sm font-medium text-gray-900 hover:bg-gray-50 sm:w-auto lg:w-full">Write a review</a>
            </div>
          </div>

          <div class="mt-16 lg:col-span-7 lg:col-start-6 lg:mt-0">
            <h3 class="sr-only">Recent reviews</h3>

            <div class="flow-root">
              <div class="-my-12 divide-y divide-gray-200">
                <div v-for="review in reviews.featured" :key="review.id" class="py-12">
                  <div class="flex items-center">
                    <img :src="review.avatarSrc" :alt="`${review.author}.`" class="size-12 rounded-full" />
                    <div class="ml-4">
                      <h4 class="text-sm font-bold text-gray-900">{{ review.author }}</h4>
                      <div class="mt-1 flex items-center">
                        <StarIcon v-for="rating in [0, 1, 2, 3, 4]" :key="rating" :class="[review.rating > rating ? 'text-yellow-400' : 'text-gray-300', 'size-5 shrink-0']" aria-hidden="true" />
                      </div>
                      <p class="sr-only">{{ review.rating }} out of 5 stars</p>
                    </div>
                  </div>

                  <div class="mt-4 space-y-6 text-base text-gray-600 italic" v-html="review.content"></div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </main>

    <footer aria-labelledby="footer-heading" class="bg-white">
      <h2 id="footer-heading" class="sr-only">Footer</h2>
      <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div class="border-t border-gray-200 py-20">
          <div class="grid grid-cols-1 md:grid-flow-col md:auto-rows-min md:grid-cols-12 md:gap-x-8 md:gap-y-16">
            <!-- Image section -->
            <div class="col-span-1 md:col-span-2 lg:col-start-1 lg:row-start-1">
              <img src="https://tailwindcss.com/plus-assets/img/logos/mark.svg?color=indigo&shade=600" alt="" class="h-8 w-auto" />
            </div>

            <!-- Sitemap sections -->
            <div class="col-span-6 mt-10 grid grid-cols-2 gap-8 sm:grid-cols-3 md:col-span-8 md:col-start-3 md:row-start-1 md:mt-0 lg:col-span-6 lg:col-start-2">
              <div class="grid grid-cols-1 gap-y-12 sm:col-span-2 sm:grid-cols-2 sm:gap-x-8">
                <div>
                  <h3 class="text-sm font-medium text-gray-900">Products</h3>
                  <ul role="list" class="mt-6 space-y-6">
                    <li v-for="item in footerNavigation.products" :key="item.name" class="text-sm">
                      <a :href="item.href" class="text-gray-500 hover:text-gray-600">{{ item.name }}</a>
                    </li>
                  </ul>
                </div>
                <div>
                  <h3 class="text-sm font-medium text-gray-900">Company</h3>
                  <ul role="list" class="mt-6 space-y-6">
                    <li v-for="item in footerNavigation.company" :key="item.name" class="text-sm">
                      <a :href="item.href" class="text-gray-500 hover:text-gray-600">{{ item.name }}</a>
                    </li>
                  </ul>
                </div>
              </div>
              <div>
                <h3 class="text-sm font-medium text-gray-900">Customer Service</h3>
                <ul role="list" class="mt-6 space-y-6">
                  <li v-for="item in footerNavigation.customerService" :key="item.name" class="text-sm">
                    <a :href="item.href" class="text-gray-500 hover:text-gray-600">{{ item.name }}</a>
                  </li>
                </ul>
              </div>
            </div>

            <!-- Newsletter section -->
            <div class="mt-12 md:col-span-8 md:col-start-3 md:row-start-2 md:mt-0 lg:col-span-4 lg:col-start-9 lg:row-start-1">
              <h3 class="text-sm font-medium text-gray-900">Sign up for our newsletter</h3>
              <p class="mt-6 text-sm text-gray-500">The latest deals and savings, sent to your inbox weekly.</p>
              <form class="mt-2 flex sm:max-w-md">
                <input id="email-address" type="text" autocomplete="email" required="" aria-label="Email address" class="block w-full rounded-md bg-white px-4 py-2 text-base text-gray-900 outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600" />
                <div class="ml-4 shrink-0">
                  <button type="submit" class="flex w-full items-center justify-center rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-base font-medium text-white shadow-xs hover:bg-indigo-700 focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:outline-hidden">Sign up</button>
                </div>
              </form>
            </div>
          </div>
        </div>

        <div class="border-t border-gray-100 py-10 text-center">
          <p class="text-sm text-gray-500">&copy; 2021 Your Company, Inc. All rights reserved.</p>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import {
  Dialog,
  DialogPanel,
  Popover,
  PopoverButton,
  PopoverGroup,
  PopoverPanel,
  Tab,
  TabGroup,
  TabList,
  TabPanel,
  TabPanels,
  TransitionChild,
  TransitionRoot,
} from '@headlessui/vue'
import { Bars3Icon, MagnifyingGlassIcon, ShieldCheckIcon, ShoppingBagIcon, XMarkIcon } from '@heroicons/vue/24/outline'
import { CheckCircleIcon, CheckIcon, QuestionMarkCircleIcon, StarIcon } from '@heroicons/vue/20/solid'

const navigation = {
  categories: [
    {
      id: 'women',
      name: 'Women',
      featured: [
        {
          name: 'New Arrivals',
          href: '#',
          imageSrc: 'https://tailwindcss.com/plus-assets/img/ecommerce-images/mega-menu-category-01.jpg',
          imageAlt: 'Models sitting back to back, wearing Basic Tee in black and bone.',
        },
        {
          name: 'Basic Tees',
          href: '#',
          imageSrc: 'https://tailwindcss.com/plus-assets/img/ecommerce-images/mega-menu-category-02.jpg',
          imageAlt: 'Close up of Basic Tee fall bundle with off-white, ochre, olive, and black tees.',
        },
      ],
      sections: [
        {
          id: 'clothing',
          name: 'Clothing',
          items: [
            { name: 'Tops', href: '#' },
            { name: 'Dresses', href: '#' },
            { name: 'Pants', href: '#' },
            { name: 'Denim', href: '#' },
            { name: 'Sweaters', href: '#' },
            { name: 'T-Shirts', href: '#' },
            { name: 'Jackets', href: '#' },
            { name: 'Activewear', href: '#' },
            { name: 'Browse All', href: '#' },
          ],
        },
        {
          id: 'accessories',
          name: 'Accessories',
          items: [
            { name: 'Watches', href: '#' },
            { name: 'Wallets', href: '#' },
            { name: 'Bags', href: '#' },
            { name: 'Sunglasses', href: '#' },
            { name: 'Hats', href: '#' },
            { name: 'Belts', href: '#' },
          ],
        },
        {
          id: 'brands',
          name: 'Brands',
          items: [
            { name: 'Full Nelson', href: '#' },
            { name: 'My Way', href: '#' },
            { name: 'Re-Arranged', href: '#' },
            { name: 'Counterfeit', href: '#' },
            { name: 'Significant Other', href: '#' },
          ],
        },
      ],
    },
    {
      id: 'men',
      name: 'Men',
      featured: [
        {
          name: 'New Arrivals',
          href: '#',
          imageSrc:
            'https://tailwindcss.com/plus-assets/img/ecommerce-images/product-page-04-detail-product-shot-01.jpg',
          imageAlt: 'Drawstring top with elastic loop closure and textured interior padding.',
        },
        {
          name: 'Artwork Tees',
          href: '#',
          imageSrc: 'https://tailwindcss.com/plus-assets/img/ecommerce-images/category-page-02-image-card-06.jpg',
          imageAlt:
            'Three shirts in gray, white, and blue arranged on table with same line drawing of hands and shapes overlapping on front of shirt.',
        },
      ],
      sections: [
        {
          id: 'clothing',
          name: 'Clothing',
          items: [
            { name: 'Tops', href: '#' },
            { name: 'Pants', href: '#' },
            { name: 'Sweaters', href: '#' },
            { name: 'T-Shirts', href: '#' },
            { name: 'Jackets', href: '#' },
            { name: 'Activewear', href: '#' },
            { name: 'Browse All', href: '#' },
          ],
        },
        {
          id: 'accessories',
          name: 'Accessories',
          items: [
            { name: 'Watches', href: '#' },
            { name: 'Wallets', href: '#' },
            { name: 'Bags', href: '#' },
            { name: 'Sunglasses', href: '#' },
            { name: 'Hats', href: '#' },
            { name: 'Belts', href: '#' },
          ],
        },
        {
          id: 'brands',
          name: 'Brands',
          items: [
            { name: 'Re-Arranged', href: '#' },
            { name: 'Counterfeit', href: '#' },
            { name: 'Full Nelson', href: '#' },
            { name: 'My Way', href: '#' },
          ],
        },
      ],
    },
  ],
  pages: [
    { name: 'Company', href: '#' },
    { name: 'Stores', href: '#' },
  ],
}
const userNavigation = [
  { name: 'Sign in', href: '#' },
  { name: 'Create account', href: '#' },
]
const product = {
  name: 'Everyday Ruck Snack',
  href: '#',
  price: '$220',
  description:
    "Don't compromise on snack-carrying capacity with this lightweight and spacious bag. The drawstring top keeps all your favorite chips, crisps, fries, biscuits, crackers, and cookies secure.",
  imageSrc: 'https://tailwindcss.com/plus-assets/img/ecommerce-images/product-page-04-featured-product-shot.jpg',
  imageAlt: 'Light green canvas bag with black straps, handle, front zipper pouch, and drawstring top.',
  breadcrumbs: [
    { id: 1, name: 'Travel', href: '#' },
    { id: 2, name: 'Bags', href: '#' },
  ],
  sizes: [
    { name: '18L', description: 'Perfect for a reasonable amount of snacks.' },
    { name: '20L', description: 'Enough room for a large amount of snacks.' },
  ],
}
const policies = [
  {
    name: 'Free delivery all year long',
    description:
      'Name another place that offers year long free delivery? We’ll be waiting. Order now and you’ll get delivery absolutely free.',
    imageSrc: 'https://tailwindcss.com/plus-assets/img/ecommerce/icons/icon-delivery-light.svg',
  },
  {
    name: '24/7 Customer Support',
    description:
      'Or so we want you to believe. In reality our chat widget is powered by a naive series of if/else statements that churn out canned responses. Guaranteed to irritate.',
    imageSrc: 'https://tailwindcss.com/plus-assets/img/ecommerce/icons/icon-chat-light.svg',
  },
  {
    name: 'Fast Shopping Cart',
    description:
      "Look at the cart in that icon, there's never been a faster cart. What does this mean for the actual checkout experience? I don't know.",
    imageSrc: 'https://tailwindcss.com/plus-assets/img/ecommerce/icons/icon-fast-checkout-light.svg',
  },
  {
    name: 'Gift Cards',
    description:
      "We sell these hoping that you will buy them for your friends and they will never actually use it. Free money for us, it's great.",
    imageSrc: 'https://tailwindcss.com/plus-assets/img/ecommerce/icons/icon-gift-card-light.svg',
  },
]
const reviews = {
  average: 4,
  totalCount: 1624,
  counts: [
    { rating: 5, count: 1019 },
    { rating: 4, count: 162 },
    { rating: 3, count: 97 },
    { rating: 2, count: 199 },
    { rating: 1, count: 147 },
  ],
  featured: [
    {
      id: 1,
      rating: 5,
      content: `
        <p>This is the bag of my dreams. I took it on my last vacation and was able to fit an absurd amount of snacks for the many long and hungry flights.</p>
      `,
      author: 'Emily Selman',
      avatarSrc:
        'https://images.unsplash.com/photo-1502685104226-ee32379fefbe?ixlib=rb-=eyJhcHBfaWQiOjEyMDd9&auto=format&fit=facearea&facepad=8&w=256&h=256&q=80',
    },
    {
      id: 2,
      rating: 5,
      content: `
        <p>Before getting the Ruck Snack, I struggled my whole life with pulverized snacks, endless crumbs, and other heartbreaking snack catastrophes. Now, I can stow my snacks with confidence and style!</p>
      `,
      author: 'Hector Gibbons',
      avatarSrc:
        'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?ixlib=rb-=eyJhcHBfaWQiOjEyMDd9&auto=format&fit=facearea&facepad=8&w=256&h=256&q=80',
    },
    {
      id: 3,
      rating: 4,
      content: `
        <p>I love how versatile this bag is. It can hold anything ranging from cookies that come in trays to cookies that come in tins.</p>
      `,
      author: 'Mark Edwards',
      avatarSrc:
        'https://images.unsplash.com/photo-1519244703995-f4e0f30006d5?ixlib=rb-1.2.1&ixqx=oilqXxSqey&ixid=eyJhcHBfaWQiOjEyMDd9&auto=format&fit=facearea&facepad=2&w=256&h=256&q=80',
    },
  ],
}
const footerNavigation = {
  products: [
    { name: 'Bags', href: '#' },
    { name: 'Tees', href: '#' },
    { name: 'Objects', href: '#' },
    { name: 'Home Goods', href: '#' },
    { name: 'Accessories', href: '#' },
  ],
  company: [
    { name: 'Who we are', href: '#' },
    { name: 'Sustainability', href: '#' },
    { name: 'Press', href: '#' },
    { name: 'Careers', href: '#' },
    { name: 'Terms & Conditions', href: '#' },
    { name: 'Privacy', href: '#' },
  ],
  customerService: [
    { name: 'Contact', href: '#' },
    { name: 'Shipping', href: '#' },
    { name: 'Returns', href: '#' },
    { name: 'Warranty', href: '#' },
    { name: 'Secure Payments', href: '#' },
    { name: 'FAQ', href: '#' },
    { name: 'Find a store', href: '#' },
  ],
}

const open = ref(false)
</script>