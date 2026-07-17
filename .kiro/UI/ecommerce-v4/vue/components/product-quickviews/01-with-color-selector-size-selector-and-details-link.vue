<template>
  <TransitionRoot as="template" :show="open">
    <Dialog class="relative z-10" @close="open = false">
      <TransitionChild as="template" enter="ease-out duration-300" enter-from="opacity-0" enter-to="" leave="ease-in duration-200" leave-from="" leave-to="opacity-0">
        <div class="fixed inset-0 hidden bg-gray-500/75 transition-opacity md:block"></div>
      </TransitionChild>

      <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
        <div class="flex min-h-full items-stretch justify-center text-center md:items-center md:px-2 lg:px-4">
          <!-- This element is to trick the browser into centering the modal contents. -->
          <span class="hidden md:inline-block md:h-screen md:align-middle" aria-hidden="true">&#8203;</span>
          <TransitionChild as="template" enter="ease-out duration-300" enter-from="opacity-0 translate-y-4 md:translate-y-0 md:scale-95" enter-to=" translate-y-0 md:scale-100" leave="ease-in duration-200" leave-from=" translate-y-0 md:scale-100" leave-to="opacity-0 translate-y-4 md:translate-y-0 md:scale-95">
            <DialogPanel class="flex w-full transform text-left text-base transition md:my-8 md:max-w-2xl md:px-4 lg:max-w-4xl">
              <div class="relative flex w-full items-center overflow-hidden bg-white px-4 pt-14 pb-8 shadow-2xl sm:px-6 sm:pt-8 md:p-6 lg:p-8">
                <button type="button" class="absolute top-4 right-4 text-gray-400 hover:text-gray-500 sm:top-8 sm:right-6 md:top-6 md:right-6 lg:top-8 lg:right-8" @click="open = false">
                  <span class="sr-only">Close</span>
                  <XMarkIcon class="size-6" aria-hidden="true" />
                </button>

                <div class="grid w-full grid-cols-1 items-start gap-x-6 gap-y-8 sm:grid-cols-12 lg:items-center lg:gap-x-8">
                  <img :src="product.imageSrc" :alt="product.imageAlt" class="aspect-2/3 w-full rounded-lg bg-gray-100 object-cover sm:col-span-4 lg:col-span-5" />
                  <div class="sm:col-span-8 lg:col-span-7">
                    <h2 class="text-xl font-medium text-gray-900 sm:pr-12">{{ product.name }}</h2>

                    <section aria-labelledby="information-heading" class="mt-1">
                      <h3 id="information-heading" class="sr-only">Product information</h3>

                      <p class="font-medium text-gray-900">{{ product.price }}</p>

                      <!-- Reviews -->
                      <div class="mt-4">
                        <h4 class="sr-only">Reviews</h4>
                        <div class="flex items-center">
                          <p class="text-sm text-gray-700">
                            {{ product.rating }}
                            <span class="sr-only"> out of 5 stars</span>
                          </p>
                          <div class="ml-1 flex items-center">
                            <StarIcon v-for="rating in [0, 1, 2, 3, 4]" :key="rating" :class="[product.rating > rating ? 'text-yellow-400' : 'text-gray-200', 'size-5 shrink-0']" aria-hidden="true" />
                          </div>
                          <div class="ml-4 hidden lg:flex lg:items-center">
                            <span class="text-gray-300" aria-hidden="true">&middot;</span>
                            <a href="#" class="ml-4 text-sm font-medium text-indigo-600 hover:text-indigo-500">See all {{ product.reviewCount }} reviews</a>
                          </div>
                        </div>
                      </div>
                    </section>

                    <section aria-labelledby="options-heading" class="mt-8">
                      <h3 id="options-heading" class="sr-only">Product options</h3>

                      <form>
                        <!-- Color picker -->
                        <fieldset aria-label="Choose a color">
                          <legend class="text-sm font-medium text-gray-900">Color</legend>

                          <div class="mt-2 flex items-center gap-x-3">
                            <div v-for="color in product.colors" :key="color.id" class="flex rounded-full outline -outline-offset-1 outline-black/10">
                              <input :aria-label="color.name" type="radio" name="color" :value="color.id" :checked="color === product.colors[0]" :class="[color.classes, 'size-8 appearance-none rounded-full forced-color-adjust-none checked:outline-2 checked:outline-offset-2 focus-visible:outline-3 focus-visible:outline-offset-3']" />
                            </div>
                          </div>
                        </fieldset>

                        <!-- Size picker -->
                        <fieldset class="mt-8" aria-label="Choose a size">
                          <div class="flex items-center justify-between">
                            <div class="text-sm font-medium text-gray-900">Size</div>
                            <a href="#" class="text-sm font-medium text-indigo-600 hover:text-indigo-500">Size guide</a>
                          </div>
                          <div class="mt-2 grid grid-cols-4 gap-3">
                            <label v-for="size in product.sizes" :key="size.id" :aria-label="size.name" class="group relative flex items-center justify-center rounded-md border border-gray-300 bg-white p-3 has-checked:border-indigo-600 has-checked:bg-indigo-600 has-focus-visible:outline-2 has-focus-visible:outline-offset-2 has-focus-visible:outline-indigo-600 has-disabled:border-gray-400 has-disabled:bg-gray-200 has-disabled:opacity-25">
                              <input type="radio" name="size" :value="size.id" :checked="size === product.sizes[2]" :disabled="!size.inStock" class="absolute inset-0 appearance-none focus:outline-none disabled:cursor-not-allowed" />
                              <span class="text-sm font-medium text-gray-900 uppercase group-has-checked:text-white">{{ size.name }}</span>
                            </label>
                          </div>
                        </fieldset>

                        <button type="submit" class="mt-8 flex w-full items-center justify-center rounded-md border border-transparent bg-indigo-600 px-8 py-3 text-base font-medium text-white hover:bg-indigo-700 focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:outline-hidden">Add to bag</button>

                        <p class="absolute top-4 left-4 text-center sm:static sm:mt-8">
                          <a :href="product.href" class="font-medium text-indigo-600 hover:text-indigo-500">View full details</a>
                        </p>
                      </form>
                    </section>
                  </div>
                </div>
              </div>
            </DialogPanel>
          </TransitionChild>
        </div>
      </div>
    </Dialog>
  </TransitionRoot>
</template>

<script setup>
import { ref } from 'vue'
import { Dialog, DialogPanel, TransitionChild, TransitionRoot } from '@headlessui/vue'
import { XMarkIcon } from '@heroicons/vue/24/outline'
import { StarIcon } from '@heroicons/vue/20/solid'

const product = {
  name: "Women's Basic Tee",
  price: '$32',
  rating: 3.9,
  reviewCount: 512,
  href: '#',
  imageSrc: 'https://tailwindcss.com/plus-assets/img/ecommerce-images/product-page-01-featured-product-shot.jpg',
  imageAlt: "Back of women's Basic Tee in black.",
  colors: [
    { id: 'black', name: 'Black', classes: 'bg-gray-900 checked:outline-gray-900' },
    { id: 'heather-grey', name: 'Heather Grey', classes: 'bg-gray-400 checked:outline-gray-400' },
  ],
  sizes: [
    { id: 'xxs', name: 'XXS', inStock: true },
    { id: 'xs', name: 'XS', inStock: true },
    { id: 's', name: 'S', inStock: true },
    { id: 'm', name: 'M', inStock: true },
    { id: 'l', name: 'L', inStock: true },
    { id: 'xl', name: 'XL', inStock: false },
  ],
}

const open = ref(false)
</script>